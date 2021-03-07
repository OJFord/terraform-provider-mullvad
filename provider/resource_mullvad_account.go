package provider

import (
	"github.com/OJFord/terraform-provider-mullvad/mullvadapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func resourceMullvadAccount() *schema.Resource {
	return &schema.Resource{
		Description: "Provides access to a Mullvad account. Required if the provider is not configured with an `account_id`.",
		Schema:      accountSchema,

		Importer: &schema.ResourceImporter{
			State: importAccount,
		},

		Create: resourceMullvadAccountCreate,
		Read:   resourceMullvadAccountRead,
		Delete: resourceMullvadAccountDelete,
	}
}

func importAccount(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}

func resourceMullvadAccountCreate(d *schema.ResourceData, m interface{}) error {
	acc, err := m.(*mullvadapi.Client).CreateAccount()
	if err != nil {
		return err
	}

	log.Printf("Created %s", acc.Token)
	populateAccountResource(d, acc)
	log.Printf("Created %s", d.Id())
	return nil
}

func resourceMullvadAccountRead(d *schema.ResourceData, m interface{}) error {
	acc, err := m.(*mullvadapi.Client).Login(d.Id())
	if err != nil {
		return err
	}

	log.Printf("Reading %s", d.Id())
	populateAccountResource(d, acc)
	return nil
}

func resourceMullvadAccountDelete(d *schema.ResourceData, m interface{}) error {
	// I don't think there's a way to delete, so just NOP & forget.
	return nil
}
