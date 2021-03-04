PLAN_FILE = "do.tfplan"

init:
	terraform get
	terraform init \
		-reconfigure \
		-input=false \
		-get=true \
		-upgrade \
		-backend=true \
		-lock=true

plan:
	terraform plan -input=false -out $(PLAN_FILE)

apply:
	terraform apply $(PLAN_FILE)
	rm -f $(PLAN_FILE)

destroy:
	terraform destroy
