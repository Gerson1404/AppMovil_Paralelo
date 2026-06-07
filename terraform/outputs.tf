output "api_gateway_url" {
  description = "La URL base pública de tu API en AWS"
  value       = aws_apigatewayv2_api.gin_api.api_endpoint
}