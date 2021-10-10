docker-build:
	bash builders/docker-build.sh
docker-push:
	bash builders/docker-build.sh push
delete-table:
	bash misc/delete-table.sh