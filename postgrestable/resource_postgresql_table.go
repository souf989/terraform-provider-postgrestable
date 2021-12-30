package postgresql

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/lib/pq"
)

func resourcePostgreSqlTable() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePostgreSqlTableCreate,
		ReadContext:   resourcePostgreSqlTableRead,
		UpdateContext: resourcePostgreSqlTableUpdate,
		DeleteContext: resourcePostgreSqlTableDelete,
		Schema: map[string]*schema.Schema{
			"table": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The PostgreSQL table to create",
			},
			"schema": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The PostgreSQL schema where create the database",
			},
			"columns": &schema.Schema{
				Type:        schema.TypeList,
				Required:    true,
				Description: "The Column list defintions",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"type": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourcePostgreSqlTableCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	db := m.(*DBConnection)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if err := createTable(db, d); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Error creating the table" + err.Error(),
			Detail:   "Unable to create the table" + err.Error(),
		})
		return diags
	}

	d.SetId(d.Get("schema").(string) + "." + d.Get("table").(string))

	resourcePostgreSqlTableRead(ctx, d, m)

	return diags

}

func resourcePostgreSqlTableUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	return diags
}

func createTable(db *DBConnection, d *schema.ResourceData) error {

	schemaName := d.Get("schema").(string)
	tableName := d.Get("table").(string)
	var completeTableName = schemaName + "." + pq.QuoteIdentifier(tableName)

	b := bytes.NewBufferString("CREATE TABLE IF NOT EXISTS ")
	fmt.Fprintln(b, completeTableName, " (")

	columns := d.Get("columns").([]interface{})

	for i, column := range columns {
		col := column.(map[string]interface{})
		fmt.Fprint(b, " ", pq.QuoteIdentifier(col["name"].(string)), " ", col["type"].(string))
		if i < len(columns)-1 {
			fmt.Fprintln(b, " ,")
		}
	}
	fmt.Fprint(b, " )")

	sql := b.String()

	if _, err := db.Exec(sql); err != nil {
		return fmt.Errorf("Error creating table %q: %w", tableName, err, sql)
	}

	var err error

	return err
}

func resourcePostgreSqlTableRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	db := m.(*DBConnection)
	return getTableColumnsDefintion(db, d)
}

func resourcePostgreSqlTableDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags

}

func getTableColumnsDefintion(db *DBConnection, d *schema.ResourceData) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	tablename := d.Get("table").(string)
	rows, err := db.Query("select column_name,udt_name from INFORMATION_SCHEMA.columns where table_name=$1 order by ordinal_position", tablename)

	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Error while looking for columns of table" + tablename,
			Detail:   "Error while looking for columns of table" + tablename,
		})
		return diags
	}

	var columns []map[string]interface{}

	var name, column_type string
	for rows.Next() {
		if err := rows.Scan(&name, &column_type); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Could not scan tables columns",
				Detail:   "Unable to scan tables columns",
			})
			return diags
		}
		column := make(map[string]interface{})
		column["name"] = name
		column["type"] = column_type
		columns = append(columns, column)
	}

	if err := d.Set("columns", columns); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Can not connect to database",
			Detail:   "Unable to connect user to the database",
		})
		return diags
	}
	return nil
}
