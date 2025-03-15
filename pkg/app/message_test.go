package app

import (
	"fmt"
	"strings"
	"testing"
)

func Test_splitMessages(t *testing.T) {
	body := ":white_check_mark: **Plan Result**\nproject: `aws` dir: `test/aws` workspace: `default`\n\n```\nPlan: 1 to add, 0 to change, 0 to destroy.\n```\n\n\n<details><summary>Show Output</summary>\n\n```diff\n\n  # aws_s3_bucket.test_s3 will be created\n+   resource \"aws_s3_bucket\" \"test_s3\" {\n+       acceleration_status         = (known after apply)\n+       acl                         = (known after apply)\n+       arn                         = (known after apply)\n+       bucket                      = \"test-mu-bucket-dev\"\n+       bucket_domain_name          = (known after apply)\n+       bucket_prefix               = (known after apply)\n+       bucket_regional_domain_name = (known after apply)\n+       force_destroy               = false\n+       hosted_zone_id              = (known after apply)\n+       id                          = (known after apply)\n+       object_lock_enabled         = (known after apply)\n+       policy                      = (known after apply)\n+       region                      = (known after apply)\n+       request_payer               = (known after apply)\n+       tags                        = {\n+           \"Environment\" = \"dev\"\n+           \"Name\"        = \"dev bucket\"\n        }\n+       tags_all                    = {\n+           \"Environment\" = \"dev\"\n+           \"Name\"        = \"dev bucket\"\n        }\n+       website_domain              = (known after apply)\n+       website_endpoint            = (known after apply)\n    }\n\nPlan: 1 to add, 0 to change, 0 to destroy.\n\nWarning: AWS account ID not found for provider\n\n  with provider[\"registry.terraform.io/hashicorp/aws\"],\n  on main.tf line 6, in provider \"aws\":\n   6: provider \"aws\" {\n\nSee\nhttps://registry.terraform.io/providers/hashicorp/aws/latest/docs#skip_requesting_account_id\nfor implications.\n\n```\n</details>\n\n**next step**\n- To apply this plan, comment:\n  ```\n  mu apply -p aws\n  ```\n- To delete this plan and lock, comment:\n  ```\n  mu unlock -p aws\n  ```\n- To plan this project again, comment:\n  ```\n  mu plan -p aws\n  ```\n> [!WARNING]\n> Warning: AWS account ID not found for provider\n> \n>   with provider[\"registry.terraform.io/hashicorp/aws\"],\n>   on main.tf line 6, in provider \"aws\":\n>    6: provider \"aws\" {\n> \n> See\n> https://registry.terraform.io/providers/hashicorp/aws/latest/docs#skip_requesting_account_id\n> for implications.\n\n\n"
	app := &App{}
	msgs := app.splitMessages(strings.NewReader(body))
	for _, msg := range msgs {
		fmt.Println(msg)
	}
}

func TestApp_helpMessage(t *testing.T) {
	app := &App{}
	msg := app.helpMessage()
	fmt.Println(msg)
}
