from django.urls import path
from rest_framework_simplejwt.views import TokenRefreshView

from apps.api.views import LoginView
from apps.api.views import CompaniesWithServicesView
from apps.api.views import CheckCardPaymentStatusView, CheckInvoicePaymentStatusView
from apps.api.views import CompanyInfoByInnView, InvoiceStatusUpdateView, InvoiceDeleteView
from apps.api.views import ServiceUsersAPIView
from apps.api.views.relay import RelayTokenValidateView

urlpatterns = [
    path('login/', LoginView.as_view(), name='login_api'),
    path('user/services/', CompaniesWithServicesView.as_view(), name='companies-with-services'),
    path('token/refresh/', TokenRefreshView.as_view(), name='token_refresh'),
    path('relay/token/validate/', RelayTokenValidateView.as_view(), name='relay_token_validate'),

    path('inn-info/', CompanyInfoByInnView.as_view(), name='inn_info'),
    path('tbank/check-payments-invoice/', CheckInvoicePaymentStatusView.as_view(), name='check_payments_invoice'),
    path('tbank/check-payments-card/', CheckCardPaymentStatusView.as_view(), name='check_payments_card'),

    path('company/invoice/change-status/', InvoiceStatusUpdateView.as_view(), name='invoice_status_update'),
    path('company/invoice/delete/', InvoiceDeleteView.as_view(), name='invoice_delete'),
    path('service/<uuid:service_id>/users/', ServiceUsersAPIView.as_view(), name='service_users'),
] 