package controllers

import (
	"context"
	"testing"

	calicov1 "github.com/projectcalico/api/pkg/apis/projectcalico/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestTenantPolicyReconciler_Reconcile(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, calicov1.AddToScheme(scheme))

	tests := []struct {
		name             string
		pods             []client.Object
		expectedPolicies []string
	}{
		{
			name: "single tenant with pods",
			pods: []client.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-1",
						Namespace: "cloudbridge",
						Labels: map[string]string{
							tenantLabel: "tenant-123",
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				},
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-2",
						Namespace: "cloudbridge",
						Labels: map[string]string{
							tenantLabel: "tenant-123",
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				},
			},
			expectedPolicies: []string{
				"tenant-tenant-123-allow-intra",
				"tenant-tenant-123-allow-dns",
				"tenant-tenant-123-default-deny",
			},
		},
		{
			name: "multiple tenants",
			pods: []client.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-1",
						Namespace: "cloudbridge",
						Labels: map[string]string{
							tenantLabel: "tenant-123",
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				},
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-2",
						Namespace: "cloudbridge",
						Labels: map[string]string{
							tenantLabel: "tenant-456",
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				},
			},
			expectedPolicies: []string{
				"tenant-tenant-123-allow-intra",
				"tenant-tenant-123-allow-dns",
				"tenant-tenant-123-default-deny",
				"tenant-tenant-456-allow-intra",
				"tenant-tenant-456-allow-dns",
				"tenant-tenant-456-default-deny",
			},
		},
		{
			name: "no tenant pods",
			pods: []client.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-1",
						Namespace: "cloudbridge",
						Labels: map[string]string{
							"app": "other-app",
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				},
			},
			expectedPolicies: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем fake client
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(tt.pods...).
				Build()

			// Создаем reconciler
			reconciler := &TenantPolicyReconciler{
				Client: fakeClient,
				Scheme: scheme,
				Log:    zap.New(zap.UseDevMode(true)),
			}

			// Запускаем reconcile
			_, err := reconciler.Reconcile(context.Background(), ctrl.Request{})
			require.NoError(t, err)

			// Проверяем созданные политики
			for _, policyName := range tt.expectedPolicies {
				var policy calicov1.GlobalNetworkPolicy
				err := fakeClient.Get(context.Background(), types.NamespacedName{Name: policyName}, &policy)
				require.NoError(t, err, "Policy %s should exist", policyName)
				assert.Equal(t, policyName, policy.Name)
			}

			// Проверяем, что нет лишних политик
			var policies calicov1.GlobalNetworkPolicyList
			err = fakeClient.List(context.Background(), &policies)
			require.NoError(t, err)
			assert.Equal(t, len(tt.expectedPolicies), len(policies.Items))
		})
	}
}

func TestBuildAllowIntraSpec(t *testing.T) {
	tenantID := "tenant-123"
	spec := buildAllowIntraSpec(tenantID)

	assert.Equal(t, 100.0, *spec.Order)
	assert.Equal(t, `tenant.cloudbridge.io/id == "tenant-123"`, spec.Selector)
	assert.Len(t, spec.Types, 2)
	assert.Contains(t, spec.Types, calicov1.PolicyTypeIngress)
	assert.Contains(t, spec.Types, calicov1.PolicyTypeEgress)
	assert.Len(t, spec.Ingress, 1)
	assert.Len(t, spec.Egress, 1)

	// Проверяем селекторы
	expectedSelector := `tenant.cloudbridge.io/id == "tenant-123"`
	assert.Equal(t, expectedSelector, spec.Ingress[0].Source.Selector)
	assert.Equal(t, expectedSelector, spec.Ingress[0].Destination.Selector)
	assert.Equal(t, expectedSelector, spec.Egress[0].Destination.Selector)
}

func TestBuildAllowDNSSpec(t *testing.T) {
	tenantID := "tenant-123"
	spec := buildAllowDNSSpec(tenantID)

	assert.Equal(t, 150.0, *spec.Order)
	assert.Equal(t, `tenant.cloudbridge.io/id == "tenant-123"`, spec.Selector)
	assert.Len(t, spec.Types, 1)
	assert.Contains(t, spec.Types, calicov1.PolicyTypeEgress)

	// Проверяем DNS правила (только egress)
	assert.Len(t, spec.Ingress, 0)
	assert.Len(t, spec.Egress, 1)

	// Проверяем селектор DNS сервиса
	assert.Equal(t, `k8s-app == "kube-dns"`, spec.Egress[0].Destination.Selector)
}

func TestBuildDefaultDenySpec(t *testing.T) {
	tenantID := "tenant-123"
	spec := buildDefaultDenySpec(tenantID)

	assert.Equal(t, 200.0, *spec.Order)
	assert.Equal(t, `tenant.cloudbridge.io/id == "tenant-123"`, spec.Selector)
	assert.Len(t, spec.Types, 2)
	assert.Contains(t, spec.Types, calicov1.PolicyTypeIngress)
	assert.Contains(t, spec.Types, calicov1.PolicyTypeEgress)

	// Проверяем deny правила
	assert.Len(t, spec.Ingress, 1)
	assert.Len(t, spec.Egress, 1)
	assert.Equal(t, calicov1.Deny, spec.Ingress[0].Action)
	assert.Equal(t, calicov1.Deny, spec.Egress[0].Action)
}

func TestCleanupOrphanedPolicies(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, calicov1.AddToScheme(scheme))

	// Создаем политики для активных и неактивных тенантов
	existingPolicies := []client.Object{
		&calicov1.GlobalNetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: "tenant-active-allow-intra",
			},
		},
		&calicov1.GlobalNetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: "tenant-active-allow-dns",
			},
		},
		&calicov1.GlobalNetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: "tenant-active-default-deny",
			},
		},
		&calicov1.GlobalNetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: "tenant-inactive-allow-intra",
			},
		},
		&calicov1.GlobalNetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: "tenant-inactive-allow-dns",
			},
		},
		&calicov1.GlobalNetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: "tenant-inactive-default-deny",
			},
		},
		&calicov1.GlobalNetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: "other-policy",
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(existingPolicies...).
		Build()

	reconciler := &TenantPolicyReconciler{
		Client: fakeClient,
		Scheme: scheme,
		Log:    zap.New(zap.UseDevMode(true)),
	}

	// Активные тенанты
	activeTenants := map[string]struct{}{
		"active": {},
	}

	// Запускаем cleanup
	err := reconciler.cleanupOrphanedPolicies(context.Background(), activeTenants)
	require.NoError(t, err)

	// Проверяем, что неактивные политики удалены
	var inactivePolicy calicov1.GlobalNetworkPolicy
	err = fakeClient.Get(context.Background(), types.NamespacedName{Name: "tenant-inactive-allow-intra"}, &inactivePolicy)
	assert.Error(t, err) // Должна быть удалена

	// Проверяем, что активные политики остались
	var activePolicy calicov1.GlobalNetworkPolicy
	err = fakeClient.Get(context.Background(), types.NamespacedName{Name: "tenant-active-allow-intra"}, &activePolicy)
	assert.NoError(t, err) // Должна остаться

	// Проверяем, что другие политики не затронуты
	var otherPolicy calicov1.GlobalNetworkPolicy
	err = fakeClient.Get(context.Background(), types.NamespacedName{Name: "other-policy"}, &otherPolicy)
	assert.NoError(t, err) // Должна остаться
}
