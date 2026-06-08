# 1. IAM Role: Permisos para que la Lambda se ejecute y escriba logs
resource "aws_iam_role" "lambda_exec_role" {
  name = "serverless_api_lambda_role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_policy" {
  role       = aws_iam_role.lambda_exec_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# 2. Función Lambda
resource "aws_lambda_function" "api_lambda" {
  function_name = "go-hexagonal-api"
  
  # GitHub Actions creará este archivo deployment.zip en la fase CI/CD
  filename         = "../deployment.zip"
  source_code_hash = filebase64sha256("../deployment.zip")

  # Para Go moderno en AWS, se usa este runtime y el handler SIEMPRE se llama "bootstrap"
  runtime = "provided.al2023"
  handler = "bootstrap"
  role    = aws_iam_role.lambda_exec_role.arn

  # Inyección de variables de entorno para tu main.go
  environment {
    variables = {
      DATABASE_URL = var.database_url
      JWT_SECRET   = var.jwt_secret
      GIN_MODE     = "release" # Apaga el debug de Gin para mayor rendimiento
    }
  }
}

# 3. API Gateway v2 (HTTP API - Más rápido y barato que REST API)
resource "aws_apigatewayv2_api" "gin_api" {
  name          = "go-hexagonal-gateway"
  protocol_type = "HTTP"
}

# 4. Integración entre API Gateway y Lambda
resource "aws_apigatewayv2_integration" "lambda_integration" {
  api_id           = aws_apigatewayv2_api.gin_api.id
  integration_type = "AWS_PROXY"

  integration_uri        = aws_lambda_function.api_lambda.invoke_arn
  payload_format_version = "1.0"
}

# 5. Ruta que atrapa todo el tráfico (Cualquier endpoint /api/...) y lo manda a Gin
resource "aws_apigatewayv2_route" "default_route" {
  api_id    = aws_apigatewayv2_api.gin_api.id
  route_key = "ANY /{proxy+}"
  target    = "integrations/${aws_apigatewayv2_integration.lambda_integration.id}"
}

# 6. Stage (Entorno de despliegue)
resource "aws_apigatewayv2_stage" "default_stage" {
  api_id      = aws_apigatewayv2_api.gin_api.id
  name        = "$default"
  auto_deploy = true
}

# 7. Permiso para que API Gateway invoque la Lambda
resource "aws_lambda_permission" "api_gw_invoke" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.api_lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.gin_api.execution_arn}/*/*"
}