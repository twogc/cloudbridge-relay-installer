import logging
from rest_framework.views import APIView
from rest_framework.response import Response
from rest_framework import status
import jwt
from django.conf import settings

logger = logging.getLogger(__name__)

class RelayTokenValidateView(APIView):
    def post(self, request):
        try:
            token = request.data.get('token')
            if not token:
                logger.error("No token provided in request")
                return Response(
                    {"error": "No token provided"},
                    status=status.HTTP_400_BAD_REQUEST
                )

            try:
                decoded = jwt.decode(
                    token,
                    settings.JWT_AUTH['JWT_SECRET_KEY'],
                    algorithms=[settings.JWT_AUTH['JWT_ALGORITHM']]
                )
                logger.info(f"Successfully decoded token: {decoded}")
            except jwt.ExpiredSignatureError:
                logger.error("Token has expired")
                return Response(
                    {"error": "Token has expired"},
                    status=status.HTTP_401_UNAUTHORIZED
                )
            except jwt.InvalidTokenError as e:
                logger.error(f"Invalid token: {str(e)}")
                return Response(
                    {"error": "Invalid token"},
                    status=status.HTTP_401_UNAUTHORIZED
                )

            # Check token type
            if decoded.get('type') != 'relay':
                logger.error(f"Invalid token type: {decoded.get('type')}")
                return Response(
                    {"error": "Invalid token type"},
                    status=status.HTTP_401_UNAUTHORIZED
                )

            # Check required fields
            required_fields = ['server_id', 'company_id', 'exp', 'iat']
            missing_fields = [field for field in required_fields if field not in decoded]
            if missing_fields:
                logger.error(f"Missing required fields: {missing_fields}")
                return Response(
                    {"error": f"Missing required fields: {', '.join(missing_fields)}"},
                    status=status.HTTP_400_BAD_REQUEST
                )

            logger.info("Token validation successful")
            return Response(
                {
                    "valid": True,
                    "server_id": decoded['server_id'],
                    "company_id": decoded['company_id']
                },
                status=status.HTTP_200_OK
            )

        except Exception as e:
            logger.error(f"Unexpected error during token validation: {str(e)}")
            return Response(
                {"error": "Internal server error"},
                status=status.HTTP_500_INTERNAL_SERVER_ERROR
            ) 