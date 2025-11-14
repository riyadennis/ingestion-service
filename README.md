# ingestion-service
![Coverage](https://img.shields.io/badge/Coverage-33.5%25-yellow)


To Bring up local minio for development run:
````
docker run -d \
-p 9000:9000 -p 9001:9001 \
--name minio \
-v ~/minio/data:/data \
-e "MINIO_ROOT_USER=admin" \
-e "MINIO_ROOT_PASSWORD=password123" \
quay.io/minio/minio server /data --console-address ":9001"
````
