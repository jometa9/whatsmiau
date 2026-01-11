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


