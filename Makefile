.PHONY: push-notify-app push-sample-app
v=latest
push-notify-app:
	aws ecr get-login-password --region ap-northeast-1 | docker login --username AWS --password-stdin 211125717884.dkr.ecr.ap-northeast-1.amazonaws.com &&\
	docker build -t push-notify-app push-notify-app/ &&\
	docker tag push-notify-app:latest 211125717884.dkr.ecr.ap-northeast-1.amazonaws.com/push-notify-app:${v} &&\
	docker push 211125717884.dkr.ecr.ap-northeast-1.amazonaws.com/push-notify-app:${v}

push-sample-app:
	aws ecr get-login-password --region ap-northeast-1 | docker login --username AWS --password-stdin 211125717884.dkr.ecr.ap-northeast-1.amazonaws.com
	docker build -t sample/sample-app/app sample-app/ &&\
	docker tag sample/sample-app/app:latest 211125717884.dkr.ecr.ap-northeast-1.amazonaws.com/sample/sample-app/app:${v} &&\
	docker push 211125717884.dkr.ecr.ap-northeast-1.amazonaws.com/sample/sample-app/app:${v}