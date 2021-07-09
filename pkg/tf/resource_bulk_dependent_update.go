package tf

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/riferrei/srclient"
)

func resourceBulkDependentSchemas() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBulkDependentSchemasCreate,
		ReadContext:   resourceBulkDependentSchemasRead,
		UpdateContext: resourceBulkDependentSchemasUpdate,
		DeleteContext: resourceBulkDependentSchemasDelete,
		Schema: map[string]*schema.Schema{
			"schemas_mapping": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subject": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"schema": {
							Type:             schema.TypeString,
							Required:         true,
							DiffSuppressFunc: compareSchemas,
						},
						"version": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"schema_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"references": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"subject": {
										Type:     schema.TypeString,
										Required: true,
									},
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"version": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func compareSchemas(k, old, new string, d *schema.ResourceData) bool {
	oldBuf := bytes.Buffer{}
	oldS := bufio.NewScanner(strings.NewReader(old))
	oldS.Split(bufio.ScanLines)
	for oldS.Scan() {
		line := bytes.TrimSpace(oldS.Bytes())
		if len(line) == 0 {
			continue
		}
		oldBuf.Write(line)
	}

	newBuf := bytes.Buffer{}
	newS := bufio.NewScanner(strings.NewReader(new))
	newS.Split(bufio.ScanLines)
	for newS.Scan() {
		line := bytes.TrimSpace(newS.Bytes())
		if len(line) == 0 {
			continue
		}
		newBuf.Write(line)
	}

	return newBuf.String() == oldBuf.String()
}

func resourceBulkDependentSchemasCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	client := m.(*srclient.SchemaRegistryClient)

	mappingsList := d.Get("schemas_mapping").([]interface{})

	schemasVersion := make(map[string]int)

	for i, mappingsListItemIntf := range mappingsList {
		item := mappingsListItemIntf.(map[string]interface{})
		subject := item["subject"].(string)
		schema := item["schema"].(string)
		references := item["references"].([]interface{})

		srRefs := []srclient.Reference{}

		for _, ref := range references {
			refItem := ref.(map[string]interface{})
			refSubject := refItem["subject"].(string)

			refEntry := srclient.Reference{Name: refSubject, Subject: refSubject, Version: schemasVersion[refSubject]}
			srRefs = append(srRefs, refEntry)
		}

		schemaResp, err := client.CreateSchemaWithArbitrarySubject(url.QueryEscape(subject), schema, srclient.Protobuf, srRefs...)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error talking to schemas registry, subject: %s - Error: %w", subject, err))
		}

		item["version"] = schemaResp.Version()
		item["schema_id"] = schemaResp.ID()

		// References
		tempRefMap := make(map[string]srclient.Reference)
		for _, ref := range schemaResp.References() {
			tempRefMap[ref.Subject] = ref
		}

		for j, ref := range references {
			refItem := ref.(map[string]interface{})
			refSubject := refItem["subject"].(string)

			refItem["name"] = tempRefMap[refSubject].Name
			refItem["version"] = tempRefMap[refSubject].Version

			references[j] = refItem
		}
		item["references"] = references

		mappingsList[i] = item

		schemasVersion[subject] = schemaResp.Version()
	}

	// only set the ID if something actually changed.
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	d.Set("schemas_mapping", mappingsList)

	return diags
}

func resourceBulkDependentSchemasRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	client := m.(*srclient.SchemaRegistryClient)

	mappingsList := d.Get("schemas_mapping").([]interface{})

	// Loop through all objects (can do this in parallel)
	for i, mappingsListItemIntf := range mappingsList {
		item := mappingsListItemIntf.(map[string]interface{})
		subject := item["subject"].(string)

		schemaResp, err := client.GetLatestSchemaWithArbitrarySubject(url.QueryEscape(subject))
		if err != nil {
			return diag.FromErr(fmt.Errorf("error while getting latest schema, subject: %s - Error: %w", subject, err))
		}

		itemMap := make(map[string]interface{})
		itemMap["subject"] = subject

		buf := bytes.Buffer{}
		s := bufio.NewScanner(strings.NewReader(schemaResp.Schema()))
		s.Split(bufio.ScanLines)
		for s.Scan() {
			line := s.Bytes()
			if len(line) == 0 {
				continue
			}
			buf.Write(line)
			buf.WriteByte('\n')
		}

		itemMap["schema"] = buf.String()
		itemMap["version"] = schemaResp.Version()
		itemMap["schema_id"] = schemaResp.ID()

		// Get the references and populate the field
		itemMap["references"] = schemaResp.References()

		mappingsList[i] = itemMap
	}

	d.Set("schemas_mapping", mappingsList)

	return diags
}

func resourceBulkDependentSchemasUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	client := m.(*srclient.SchemaRegistryClient)

	mappingsList := d.Get("schemas_mapping").([]interface{})

	schemasVersion := make(map[string]int)

	for i, mappingsListItemIntf := range mappingsList {
		item := mappingsListItemIntf.(map[string]interface{})
		subject := item["subject"].(string)
		schema := item["schema"].(string)
		references := item["references"].([]interface{})

		srRefs := []srclient.Reference{}

		for _, ref := range references {
			refItem := ref.(map[string]interface{})
			refSubject := refItem["subject"].(string)

			refEntry := srclient.Reference{Name: refSubject, Subject: refSubject, Version: schemasVersion[refSubject]}
			srRefs = append(srRefs, refEntry)
		}

		schemaResp, err := client.CreateSchemaWithArbitrarySubject(url.QueryEscape(subject), schema, srclient.Protobuf, srRefs...)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error talking to schemas registry, subject: %s - Error: %w", subject, err))
		}

		item["version"] = schemaResp.Version()
		item["schema_id"] = schemaResp.ID()

		// References
		tempRefMap := make(map[string]srclient.Reference)
		for _, ref := range schemaResp.References() {
			tempRefMap[ref.Subject] = ref
		}

		for j, ref := range references {
			refItem := ref.(map[string]interface{})
			refSubject := refItem["subject"].(string)

			refItem["name"] = tempRefMap[refSubject].Name
			refItem["version"] = tempRefMap[refSubject].Version

			references[j] = refItem
		}
		item["references"] = references

		mappingsList[i] = item

		schemasVersion[subject] = schemaResp.Version()
	}

	// only set the ID if something actually changed.
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	d.Set("schemas_mapping", mappingsList)

	return diags
}

func resourceBulkDependentSchemasDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	client := m.(*srclient.SchemaRegistryClient)

	mappingsList := d.Get("schemas_mapping").([]interface{})

	// Loop through all objects (in reverse because of the dependencies)
	for i := len(mappingsList) - 1; i >= 0; i-- {
		mappingsListItemIntf := mappingsList[i]
		item := mappingsListItemIntf.(map[string]interface{})
		subject := item["subject"].(string)

		err := client.DeleteSubject(url.QueryEscape(subject), false)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error while deleting subject, subject: %s - Error: %w", subject, err))
		}
	}

	return diags
}
