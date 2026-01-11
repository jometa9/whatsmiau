envio de eventos webhook
que pueda hacer el ejecutable y correrlo en cuarquier lado



=============================================================

Proceso para traer actualizaciones del repositorio original y mergearlas a tu rama:

Asegúrate de estar en tu rama (ya estás en pingcore-features):

git checkout pingcore-features

Trae las actualizaciones del repositorio original:

git fetch upstream

Mergea los cambios de upstream/master a tu rama:

git merge upstream/master

Si hay conflictos, resuélvelos y luego:

git add .git commit -m "Merge upstream/master into pingcore-features"

Alternativa: usar rebase en lugar de merge (mantiene un historial más limpio):

git rebase upstream/master




# 1. Crear instancia
curl -X POST http://localhost:8080/v1/instance \
  -H "Content-Type: application/json" \
  -d '{"instanceName": "test-instance"}'

# 2. Conectar y obtener QR (base64)
curl -X POST http://localhost:8080/v1/instance/test-instance/connect \
  -H "Content-Type: application/json"

# 3. Obtener QR como imagen
curl -X GET http://localhost:8080/v1/instance/connect/test-instance/image \
  -o qr-code.png

# 4. Verificar estado
curl -X GET http://localhost:8080/v1/instance/test-instance/status


curl -X PUT 'http://localhost:8080/v1/instance/update/test-instance' \
  -H 'Content-Type: application/json' \
  -d '{
    "webhook": {
      "url": "https://webhookapp.dev/webhook/15de2279-53ed-4106-9d02-8894bfbba208"
    }
  }'