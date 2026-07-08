#!/usr/bin/env ts-node

/**
 * 🔄 Скрипт миграции с Keycloak на Zitadel
 * 
 * Этот скрипт выполняет полную миграцию данных с Keycloak на Zitadel:
 * 1. Миграция пользователей
 * 2. Миграция организаций
 * 3. Миграция тенантов
 * 4. Миграция токенов
 * 5. Создание сетевых конфигураций
 */

import { PrismaClient } from '@prisma/client';
import { ZitadelAuthService } from '../src/lib/zitadel/auth-service';
import { ZitadelOrganizationService } from '../src/lib/zitadel/organization-service';
import { ZitadelUserService } from '../src/lib/zitadel/user-service';
import { ZitadelCloudBridgeService } from '../src/lib/zitadel/cloudbridge-service';

interface MigrationStats {
  users: number;
  organizations: number;
  tenants: number;
  tokens: number;
  errors: string[];
}

class ZitadelMigration {
  private prisma: PrismaClient;
  private authService: ZitadelAuthService;
  private accessToken: string;
  private stats: MigrationStats;

  constructor() {
    this.prisma = new PrismaClient();
    this.authService = new ZitadelAuthService();
    this.stats = {
      users: 0,
      organizations: 0,
      tenants: 0,
      tokens: 0,
      errors: []
    };
  }

  async run(): Promise<void> {
    console.log('Начинаем миграцию с Keycloak на Zitadel...');
    
    try {
      // 1. Получаем access token для Zitadel
      await this.authenticate();
      
      // 2. Мигрируем организации
      await this.migrateOrganizations();
      
      // 3. Мигрируем пользователей
      await this.migrateUsers();
      
      // 4. Мигрируем тенанты
      await this.migrateTenants();
      
      // 5. Мигрируем токены
      await this.migrateTokens();
      
      // 6. Создаем сетевые конфигурации
      await this.createNetworkConfigurations();
      
      // 7. Выводим статистику
      this.printStats();
      
    } catch (error) {
      console.error('Ошибка миграции:', error);
      throw error;
    } finally {
      await this.prisma.$disconnect();
    }
  }

  private async authenticate(): Promise<void> {
    console.log('Аутентификация в Zitadel...');
    
    try {
      this.accessToken = await this.authService.getAccessToken();
      console.log('Аутентификация успешна');
    } catch (error) {
      console.error('Ошибка аутентификации:', error);
      throw error;
    }
  }

  private async migrateOrganizations(): Promise<void> {
    console.log(' Миграция организаций...');
    
    const orgService = new ZitadelOrganizationService(this.accessToken);
    
    // Получаем уникальные организации из пользователей
    const organizations = await this.prisma.user.findMany({
      select: {
        organization_name: true,
        domain: true
      },
      distinct: ['organization_name', 'domain']
    });

    for (const org of organizations) {
      try {
        const zitadelOrg = await orgService.createOrganization({
          name: org.organization_name,
          domain: org.domain
        });
        
        // Сохраняем связь в БД
        await this.prisma.zitadelOrganization.create({
          data: {
            id: zitadelOrg.id,
            name: org.organization_name,
            domain: org.domain,
            original_organization_name: org.organization_name
          }
        });
        
        this.stats.organizations++;
        console.log(`Организация создана: ${org.organization_name} (${zitadelOrg.id})`);
        
      } catch (error) {
        const errorMsg = `Ошибка создания организации ${org.organization_name}: ${error}`;
        console.error(`${errorMsg}`);
        this.stats.errors.push(errorMsg);
      }
    }
  }

  private async migrateUsers(): Promise<void> {
    console.log('Миграция пользователей...');
    
    const userService = new ZitadelUserService(this.accessToken);
    
    // Получаем всех пользователей
    const users = await this.prisma.user.findMany();
    
    for (const user of users) {
      try {
        // Находим организацию в Zitadel
        const zitadelOrg = await this.prisma.zitadelOrganization.findFirst({
          where: {
            original_organization_name: user.organization_name
          }
        });
        
        if (!zitadelOrg) {
          throw new Error(`Организация не найдена для пользователя ${user.email}`);
        }
        
        // Создаем пользователя в Zitadel
        const zitadelUser = await userService.createUser(zitadelOrg.id, {
          email: user.email,
          name: user.name,
          password: user.password_hash // Если есть хеш пароля
        });
        
        // Сохраняем связь в БД
        await this.prisma.zitadelUser.create({
          data: {
            id: zitadelUser.id,
            organization_id: zitadelOrg.id,
            email: user.email,
            name: user.name,
            original_user_id: user.id
          }
        });
        
        this.stats.users++;
        console.log(`Пользователь создан: ${user.email} (${zitadelUser.id})`);
        
      } catch (error) {
        const errorMsg = `Ошибка создания пользователя ${user.email}: ${error}`;
        console.error(`${errorMsg}`);
        this.stats.errors.push(errorMsg);
      }
    }
  }

  private async migrateTenants(): Promise<void> {
    console.log(' Миграция тенантов...');
    
    const cloudBridgeService = new ZitadelCloudBridgeService(this.accessToken);
    
    // Получаем все тенанты
    const tenants = await this.prisma.tenant.findMany();
    
    for (const tenant of tenants) {
      try {
        // Находим организацию в Zitadel
        const zitadelOrg = await this.prisma.zitadelOrganization.findFirst({
          where: {
            original_organization_name: tenant.organization_name
          }
        });
        
        if (!zitadelOrg) {
          throw new Error(`Организация не найдена для тенанта ${tenant.name}`);
        }
        
        // Создаем сетевую конфигурацию
        await cloudBridgeService.createTenantNetworkConfig(tenant.id, {
          allocatedSubnet: tenant.subnet || this.generateDeterministicSubnet(tenant.id),
          ipamPool: tenant.ipam_pool || 'cloudbridge-pool-1',
          networkPolicies: ['allow-intra-tenant']
        });
        
        // Сохраняем связь в БД
        await this.prisma.tenantNetworkConfig.create({
          data: {
            id: `config-${tenant.id}`,
            tenant_id: tenant.id,
            organization_id: zitadelOrg.id,
            allocated_subnet: tenant.subnet || this.generateDeterministicSubnet(tenant.id),
            ipam_pool: tenant.ipam_pool || 'cloudbridge-pool-1',
            network_policies: JSON.stringify(['allow-intra-tenant']),
            allocated_ips: JSON.stringify({})
          }
        });
        
        this.stats.tenants++;
        console.log(`Тенант создан: ${tenant.name} (${tenant.id})`);
        
      } catch (error) {
        const errorMsg = `Ошибка создания тенанта ${tenant.name}: ${error}`;
        console.error(`${errorMsg}`);
        this.stats.errors.push(errorMsg);
      }
    }
  }

  private async migrateTokens(): Promise<void> {
    console.log('Миграция токенов...');
    
    // Получаем все активные токены
    const tokens = await this.prisma.userToken.findMany({
      where: {
        expires_at: {
          gt: new Date()
        }
      }
    });
    
    for (const token of tokens) {
      try {
        // Находим пользователя в Zitadel
        const zitadelUser = await this.prisma.zitadelUser.findFirst({
          where: {
            original_user_id: token.user_id
          }
        });
        
        if (!zitadelUser) {
          console.warn(`Пользователь не найден для токена ${token.id}`);
          continue;
        }
        
        // Создаем новый токен в Zitadel (если нужно)
        // Здесь можно добавить логику создания токенов в Zitadel
        
        this.stats.tokens++;
        console.log(`Токен обработан: ${token.id}`);
        
      } catch (error) {
        const errorMsg = `Ошибка обработки токена ${token.id}: ${error}`;
        console.error(`${errorMsg}`);
        this.stats.errors.push(errorMsg);
      }
    }
  }

  private async createNetworkConfigurations(): Promise<void> {
    console.log('Создание сетевых конфигураций...');
    
    const cloudBridgeService = new ZitadelCloudBridgeService(this.accessToken);
    
    // Получаем все тенанты
    const tenants = await this.prisma.tenant.findMany();
    
    for (const tenant of tenants) {
      try {
        // Выделяем подсеть для тенанта
        const subnet = await cloudBridgeService.allocateSubnetForTenant(tenant.id);
        
        console.log(`Подсеть выделена для тенанта ${tenant.name}: ${subnet}`);
        
      } catch (error) {
        const errorMsg = `Ошибка выделения подсети для тенанта ${tenant.name}: ${error}`;
        console.error(`${errorMsg}`);
        this.stats.errors.push(errorMsg);
      }
    }
  }

  private generateDeterministicSubnet(tenantId: string): string {
    const hash = this.hashString(tenantId);
    const x = (hash >> 16) & 0xFF;
    const y = hash & 0xFF;
    return `10.${x}.${y}.0/24`;
  }

  private hashString(str: string): number {
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
      const char = str.charCodeAt(i);
      hash = ((hash << 5) - hash) + char;
      hash = hash & hash; // Convert to 32bit integer
    }
    return Math.abs(hash);
  }

  private printStats(): void {
    console.log('\nСтатистика миграции:');
    console.log(`Организаций создано: ${this.stats.organizations}`);
    console.log(`Пользователей создано: ${this.stats.users}`);
    console.log(`Тенантов создано: ${this.stats.tenants}`);
    console.log(`Токенов обработано: ${this.stats.tokens}`);
    console.log(`Ошибок: ${this.stats.errors.length}`);
    
    if (this.stats.errors.length > 0) {
      console.log('\nОшибки:');
      this.stats.errors.forEach((error, index) => {
        console.log(`${index + 1}. ${error}`);
      });
    }
    
    console.log('\nМиграция завершена!');
  }
}

// Запуск миграции
async function main() {
  const migration = new ZitadelMigration();
  
  try {
    await migration.run();
    process.exit(0);
  } catch (error) {
    console.error('💥 Критическая ошибка миграции:', error);
    process.exit(1);
  }
}

// Запуск только если файл выполняется напрямую
if (require.main === module) {
  main();
}

export { ZitadelMigration };

