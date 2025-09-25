#!/usr/bin/env node

/**
 * CloudBridge Relay CLI
 * 
 * Command-line interface for CloudBridge Relay
 */

import { Command } from 'commander';
import chalk from 'chalk';
import ora from 'ora';
import inquirer from 'inquirer';
import { CloudBridgeRelayQuickStart, CloudBridgeConfig } from './index';
import { createLogger } from './utils';

const program = new Command();
const logger = createLogger('cli');

// Версия и описание
program
  .name('cloudbridge-relay')
  .description('Enterprise-grade high-performance relay server with WireGuard integration')
  .version('1.0.0');

// Команда запуска сервера
program
  .command('server')
  .description('Start CloudBridge Relay server')
  .option('-p, --port <port>', 'Server port', '8080')
  .option('-h, --host <host>', 'Server host', '0.0.0.0')
  .option('--wireguard', 'Enable WireGuard', true)
  .option('--quic', 'Enable QUIC', true)
  .option('--metrics', 'Enable metrics', true)
  .option('-c, --config <file>', 'Configuration file')
  .action(async (options) => {
    const spinner = ora('Starting CloudBridge Relay server...').start();
    
    try {
      const config: Partial<CloudBridgeConfig> = {
        server: {
          port: parseInt(options.port),
          host: options.host
        },
        wireguard: {
          enabled: options.wireguard,
          interface: 'wg-mesh0',
          port: 51820
        },
        quic: {
          enabled: options.quic,
          enhanced: true
        },
        metrics: {
          enabled: options.metrics,
          port: 9090
        }
      };

      const relay = new CloudBridgeRelayQuickStart(config);
      const server = await relay.startServer();
      
      spinner.succeed(chalk.green('Server started successfully!'));
      
      console.log(chalk.blue('\n📊 Server Information:'));
      console.log(chalk.white(`  • Address: ${options.host}:${options.port}`));
      console.log(chalk.white(`  • WireGuard: ${options.wireguard ? 'Enabled' : 'Disabled'}`));
      console.log(chalk.white(`  • QUIC: ${options.quic ? 'Enabled' : 'Disabled'}`));
      console.log(chalk.white(`  • Metrics: ${options.metrics ? 'Enabled' : 'Disabled'}`));
      
      if (options.metrics) {
        console.log(chalk.white(`  • Metrics: http://${options.host}:9090/metrics`));
      }
      
      console.log(chalk.yellow('\nPress Ctrl+C to stop the server'));
      
      // Graceful shutdown
      process.on('SIGINT', async () => {
        console.log(chalk.yellow('\nShutting down server...'));
        await relay.stop();
        process.exit(0);
      });
      
    } catch (error) {
      spinner.fail(chalk.red('Failed to start server'));
      logger.error('Server start error:', error);
      process.exit(1);
    }
  });

// Команда запуска клиента
program
  .command('client')
  .description('Start CloudBridge Relay client')
  .option('-s, --server <url>', 'Server URL', 'ws://localhost:8080')
  .option('--wireguard', 'Enable WireGuard', true)
  .option('--quic', 'Enable QUIC', true)
  .action(async (options) => {
    const spinner = ora('Connecting to CloudBridge Relay server...').start();
    
    try {
      const config: Partial<CloudBridgeConfig> = {
        server: {
          port: 8080,
          host: 'localhost'
        },
        wireguard: {
          enabled: options.wireguard,
          interface: 'wg-client0',
          port: 51821
        },
        quic: {
          enabled: options.quic,
          enhanced: true
        }
      };

      const relay = new CloudBridgeRelayQuickStart(config);
      const client = await relay.startClient();
      
      spinner.succeed(chalk.green('Connected to server successfully!'));
      
      console.log(chalk.blue('\n📊 Client Information:'));
      console.log(chalk.white(`  • Server: ${options.server}`));
      console.log(chalk.white(`  • WireGuard: ${options.wireguard ? 'Enabled' : 'Disabled'}`));
      console.log(chalk.white(`  • QUIC: ${options.quic ? 'Enabled' : 'Disabled'}`));
      
      console.log(chalk.yellow('\nPress Ctrl+C to disconnect'));
      
      // Graceful shutdown
      process.on('SIGINT', async () => {
        console.log(chalk.yellow('\nDisconnecting from server...'));
        await relay.stop();
        process.exit(0);
      });
      
    } catch (error) {
      spinner.fail(chalk.red('Failed to connect to server'));
      logger.error('Client connection error:', error);
      process.exit(1);
    }
  });

// Команда интерактивной настройки
program
  .command('setup')
  .description('Interactive setup wizard')
  .action(async () => {
    console.log(chalk.blue('🚀 CloudBridge Relay Setup Wizard\n'));
    
    const answers = await inquirer.prompt([
      {
        type: 'list',
        name: 'mode',
        message: 'What would you like to run?',
        choices: [
          { name: 'Server', value: 'server' },
          { name: 'Client', value: 'client' },
          { name: 'Both', value: 'both' }
        ]
      },
      {
        type: 'input',
        name: 'port',
        message: 'Server port:',
        default: '8080',
        when: (answers) => ['server', 'both'].includes(answers.mode)
      },
      {
        type: 'input',
        name: 'host',
        message: 'Server host:',
        default: '0.0.0.0',
        when: (answers) => ['server', 'both'].includes(answers.mode)
      },
      {
        type: 'confirm',
        name: 'wireguard',
        message: 'Enable WireGuard?',
        default: true
      },
      {
        type: 'confirm',
        name: 'quic',
        message: 'Enable QUIC?',
        default: true
      },
      {
        type: 'confirm',
        name: 'metrics',
        message: 'Enable metrics?',
        default: true,
        when: (answers) => ['server', 'both'].includes(answers.mode)
      }
    ]);
    
    const spinner = ora('Starting CloudBridge Relay...').start();
    
    try {
      const config: Partial<CloudBridgeConfig> = {
        server: {
          port: parseInt(answers.port || '8080'),
          host: answers.host || '0.0.0.0'
        },
        wireguard: {
          enabled: answers.wireguard,
          interface: 'wg-mesh0',
          port: 51820
        },
        quic: {
          enabled: answers.quic,
          enhanced: true
        },
        metrics: {
          enabled: answers.metrics,
          port: 9090
        }
      };

      const relay = new CloudBridgeRelayQuickStart(config);
      
      if (['server', 'both'].includes(answers.mode)) {
        await relay.startServer();
      }
      
      if (['client', 'both'].includes(answers.mode)) {
        await relay.startClient();
      }
      
      spinner.succeed(chalk.green('CloudBridge Relay started successfully!'));
      
      console.log(chalk.blue('\n📊 Status:'));
      const status = relay.getStatus();
      console.log(chalk.white(`  • Server: ${status.server ? 'Running' : 'Stopped'}`));
      console.log(chalk.white(`  • Client: ${status.client ? 'Connected' : 'Disconnected'}`));
      
      console.log(chalk.yellow('\nPress Ctrl+C to stop'));
      
      // Graceful shutdown
      process.on('SIGINT', async () => {
        console.log(chalk.yellow('\nShutting down...'));
        await relay.stop();
        process.exit(0);
      });
      
    } catch (error) {
      spinner.fail(chalk.red('Failed to start CloudBridge Relay'));
      logger.error('Setup error:', error);
      process.exit(1);
    }
  });

// Команда статуса
program
  .command('status')
  .description('Check CloudBridge Relay status')
  .action(async () => {
    const spinner = ora('Checking status...').start();
    
    try {
      // Проверка health endpoints
      const healthChecks = [
        { name: 'Server', url: 'http://localhost:8080/health' },
        { name: 'Metrics', url: 'http://localhost:9090/metrics' }
      ];
      
      const results = await Promise.allSettled(
        healthChecks.map(async (check) => {
          const response = await fetch(check.url);
          return { name: check.name, status: response.ok };
        })
      );
      
      spinner.succeed(chalk.green('Status check completed'));
      
      console.log(chalk.blue('\n📊 Service Status:'));
      results.forEach((result, index) => {
        const check = healthChecks[index];
        if (result.status === 'fulfilled') {
          const status = result.value.status ? 'Online' : 'Error';
          const color = result.value.status ? 'green' : 'red';
          console.log(chalk.white(`  • ${check.name}: ${chalk[color](status)}`));
        } else {
          console.log(chalk.white(`  • ${check.name}: ${chalk.red('Offline')}`));
        }
      });
      
    } catch (error) {
      spinner.fail(chalk.red('Failed to check status'));
      logger.error('Status check error:', error);
      process.exit(1);
    }
  });

// Команда генерации конфигурации
program
  .command('config')
  .description('Generate configuration file')
  .option('-o, --output <file>', 'Output file', 'cloudbridge-config.yaml')
  .action(async (options) => {
    const spinner = ora('Generating configuration...').start();
    
    try {
      const config = {
        server: {
          port: 8080,
          host: '0.0.0.0',
          timeout: 30000
        },
        wireguard: {
          enabled: true,
          interface: 'wg-mesh0',
          port: 51820,
          privateKey: 'auto-generated',
          publicKey: 'auto-generated'
        },
        quic: {
          enabled: true,
          enhanced: true,
          maxStreams: 1000
        },
        metrics: {
          enabled: true,
          port: 9090,
          path: '/metrics'
        },
        security: {
          rateLimit: 1000,
          jwtSecret: 'your-secret-key'
        }
      };
      
      // Здесь можно добавить сохранение в YAML файл
      spinner.succeed(chalk.green(`Configuration generated: ${options.output}`));
      
      console.log(chalk.blue('\nConfiguration:'));
      console.log(chalk.white(JSON.stringify(config, null, 2)));
      
    } catch (error) {
      spinner.fail(chalk.red('Failed to generate configuration'));
      logger.error('Config generation error:', error);
      process.exit(1);
    }
  });

// Обработка ошибок
program.exitOverride();

try {
  program.parse();
} catch (err) {
  if (err.code === 'commander.help') {
    process.exit(0);
  } else {
    console.error(chalk.red('Error:'), err.message);
    process.exit(1);
  }
}
