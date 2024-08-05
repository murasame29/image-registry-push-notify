.PHONY: push-notify-app push-sample-app
v=latest
push-notify-app:
	docker build -t murasame29/image-updater:${v} push-notify-app/ &&\
	docker push murasame29/image-updater:${v}

push-sample-app:
	aws ecr get-login-password --region ap-northeast-1 | docker login --username AWS --password-stdin 211125717884.dkr.ecr.ap-northeast-1.amazonaws.com
	docker build -t sample/sample-app/app sample-app/ &&\
	docker tag sample/sample-app/app:latest 211125717884.dkr.ecr.ap-northeast-1.amazonaws.com/sample/sample-app/app:${v} &&\
	docker push 211125717884.dkr.ecr.ap-northeast-1.amazonaws.com/sample/sample-app/app:${v}