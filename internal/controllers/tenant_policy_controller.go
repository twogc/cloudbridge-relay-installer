package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	calicov1 "github.com/projectcalico/api/pkg/apis/projectcalico/v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	tenantLabel = "tenant.cloudbridge.io/id"
)

// escapeTenantID экранирует специальные символы в tenantID для безопасного использования в селекторах
func escapeTenantID(tenantID string) string {
	// Экранируем обратный слеш и кавычки
	id := strings.ReplaceAll(tenantID, `\`, `\\`)
	id = strings.ReplaceAll(id, `"`, `\"`)
	return id
}

// TenantPolicyReconciler reconciles tenant policies
type TenantPolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

// +kubebuilder:rbac:groups=projectcalico.org,resources=globalnetworkpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

func (r *TenantPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("tenantpolicy", req.NamespacedName)

	// Селектор: label exists (исправлено!)
	reqExists, err := labels.NewRequirement(tenantLabel, selection.Exists, nil)
	if err != nil {
		return ctrl.Result{}, err
	}
	selector := labels.NewSelector().Add(*reqExists)

	pods := &corev1.PodList{}
	if err := r.List(ctx, pods, &client.ListOptions{
		LabelSelector: selector,
		Namespace:     "", // кластер-скоуп
	}); err != nil {
		log.Error(err, "unable to list pods")
		return ctrl.Result{}, err
	}

	// Собираем уникальные tenant ID
	tenantIDs := make(map[string]struct{})
	for _, pod := range pods.Items {
		if v := pod.Labels[tenantLabel]; v != "" {
			tenantIDs[v] = struct{}{}
		}
	}

	log.Info("Found active tenants", "count", len(tenantIDs), "tenants", tenantIDs)

	// reconcile для каждого тенанта
	for tenantID := range tenantIDs {
		if err := r.reconcileTenant(ctx, tenantID); err != nil {
			log.Error(err, "reconcile tenant failed", "tenantID", tenantID)
			return ctrl.Result{}, err
		}
	}

	// cleanup orphaned policies
	if err := r.cleanupOrphanedPolicies(ctx, tenantIDs); err != nil {
		log.Error(err, "cleanup orphaned policies failed")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *TenantPolicyReconciler) reconcileTenant(ctx context.Context, tenantID string) error {
	// allow-intra
	if err := r.createOrUpdateGNP(ctx,
		fmt.Sprintf("tenant-%s-allow-intra", tenantID),
		buildAllowIntraSpec(tenantID),
	); err != nil {
		return err
	}

	// allow-dns (опционально)
	if err := r.createOrUpdateGNP(ctx,
		fmt.Sprintf("tenant-%s-allow-dns", tenantID),
		buildAllowDNSSpec(tenantID),
	); err != nil {
		return err
	}

	// default-deny
	if err := r.createOrUpdateGNP(ctx,
		fmt.Sprintf("tenant-%s-default-deny", tenantID),
		buildDefaultDenySpec(tenantID),
	); err != nil {
		return err
	}

	return nil
}

func (r *TenantPolicyReconciler) createOrUpdateGNP(ctx context.Context, name string, spec calicov1.GlobalNetworkPolicySpec) error {
	obj := &calicov1.GlobalNetworkPolicy{}
	obj.SetName(name)

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, obj, func() error {
		obj.Spec = spec
		// Примечание: ownerRef на Pod/NS для GNP ставить нельзя; оставляем без owner'а
		return nil
	})

	if err != nil {
		// если CRD ещё не прогрет, вернётся NotFound на список – пробросим
		return err
	}
	return nil
}

func buildAllowIntraSpec(tenantID string) calicov1.GlobalNetworkPolicySpec {
	order := 100.0
	sel := fmt.Sprintf(`%s == "%s"`, tenantLabel, escapeTenantID(tenantID))
	return calicov1.GlobalNetworkPolicySpec{
		Order:    &order,
		Selector: sel, // Политика применяется только к подам этого тенанта
		Types:    []calicov1.PolicyType{calicov1.PolicyTypeIngress, calicov1.PolicyTypeEgress},
		Ingress: []calicov1.Rule{{
			Action: calicov1.Allow,
			Source: calicov1.EntityRule{Selector: sel},
			Destination: calicov1.EntityRule{
				Selector: sel,
			},
		}},
		Egress: []calicov1.Rule{{
			Action:      calicov1.Allow,
			Destination: calicov1.EntityRule{Selector: sel},
		}},
	}
}

func buildAllowDNSSpec(tenantID string) calicov1.GlobalNetworkPolicySpec {
	order := 150.0
	selTenant := fmt.Sprintf(`%s == "%s"`, tenantLabel, escapeTenantID(tenantID))
	dnsSel := `k8s-app == "kube-dns"`

	return calicov1.GlobalNetworkPolicySpec{
		Order:    &order,
		Selector: selTenant,                                        // Политика применяется к подам тенанта
		Types:    []calicov1.PolicyType{calicov1.PolicyTypeEgress}, // Только egress для DNS
		Egress: []calicov1.Rule{
			{
				Action: calicov1.Allow,
				Destination: calicov1.EntityRule{
					Selector: dnsSel,
					// Разрешаем доступ к kube-dns (UDP/TCP порт 53)
					// Calico автоматически обработает DNS трафик
				},
			},
		},
	}
}

func buildDefaultDenySpec(tenantID string) calicov1.GlobalNetworkPolicySpec {
	order := 200.0
	sel := fmt.Sprintf(`%s == "%s"`, tenantLabel, escapeTenantID(tenantID))
	return calicov1.GlobalNetworkPolicySpec{
		Order:    &order,
		Selector: sel,
		Types:    []calicov1.PolicyType{calicov1.PolicyTypeIngress, calicov1.PolicyTypeEgress},
		Ingress:  []calicov1.Rule{{Action: calicov1.Deny}},
		Egress:   []calicov1.Rule{{Action: calicov1.Deny}},
	}
}

func (r *TenantPolicyReconciler) cleanupOrphanedPolicies(ctx context.Context, active map[string]struct{}) error {
	policies := &calicov1.GlobalNetworkPolicyList{}
	if err := r.List(ctx, policies); err != nil {
		return err
	}

	// Исправлено: удаляем по индексу, избегая "адрес переменной цикла"
	for i := range policies.Items {
		p := &policies.Items[i]
		name := p.GetName()

		// Мы удаляем только свои шаблоны
		switch {
		case strings.HasPrefix(name, "tenant-") && strings.HasSuffix(name, "-allow-intra"):
			tenantID := strings.TrimPrefix(strings.TrimSuffix(name, "-allow-intra"), "tenant-")
			if _, ok := active[tenantID]; !ok {
				if err := r.Delete(ctx, p); err != nil && !errors.IsNotFound(err) {
					r.Log.Error(err, "failed to delete orphaned allow policy", "policy", name)
				}
			}
		case strings.HasPrefix(name, "tenant-") && strings.HasSuffix(name, "-allow-dns"):
			tenantID := strings.TrimPrefix(strings.TrimSuffix(name, "-allow-dns"), "tenant-")
			if _, ok := active[tenantID]; !ok {
				if err := r.Delete(ctx, p); err != nil && !errors.IsNotFound(err) {
					r.Log.Error(err, "failed to delete orphaned dns policy", "policy", name)
				}
			}
		case strings.HasPrefix(name, "tenant-") && strings.HasSuffix(name, "-default-deny"):
			tenantID := strings.TrimPrefix(strings.TrimSuffix(name, "-default-deny"), "tenant-")
			if _, ok := active[tenantID]; !ok {
				if err := r.Delete(ctx, p); err != nil && !errors.IsNotFound(err) {
					r.Log.Error(err, "failed to delete orphaned deny policy", "policy", name)
				}
			}
		}
	}
	return nil
}

func (r *TenantPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// В main обязательно: calicov1.AddToScheme(mgr.GetScheme())

	// Предикат для фильтрации событий - реагируем только на поды с лейблом тенанта
	hasTenant := predicate.NewPredicateFuncs(func(o client.Object) bool {
		if p, ok := o.(*corev1.Pod); ok {
			_, ok := p.Labels[tenantLabel]
			return ok
		}
		return false
	})

	// Добавляем индексатор по лейблу тенанта для ускорения выборки
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Pod{}, tenantLabel, func(o client.Object) []string {
		p := o.(*corev1.Pod)
		if v, ok := p.Labels[tenantLabel]; ok && v != "" {
			return []string{v}
		}
		return nil
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		WithEventFilter(hasTenant).
		Complete(r)
}
