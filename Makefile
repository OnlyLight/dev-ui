# make check WORKFLOW_FILE="aaa.yaml"

SECRET_FILE := my.secrets
VARIABLES_FILE := my.variables
WORKFLOW_FILE := 1_wf_build.yaml

check:
	act --secret-file $(SECRET_FILE) --var-file $(VARIABLES_FILE) -W ".github/workflows/$(WORKFLOW_FILE)" -p=false

.PHONY: check
