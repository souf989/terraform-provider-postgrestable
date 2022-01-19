package postgresql

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/lib/pq"
	"regexp"
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
				Description: "The Column list definitions",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"type": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile("^[a-z]+$"), "Use only alphabetical characters"),
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

	d.SetId(d.Get("schema").(string) + "." + d.Get("table").(string) + "." + uuid.New().String())

	resourcePostgreSqlTableRead(ctx, d, m)

	return diags

}

func resourcePostgreSqlTableUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	db := m.(*DBConnection)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if d.HasChange("table") {
		if err := alterTableName(db, d); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Error update the table name ",
				Detail:   "Unable to update the name of the table" + err.Error(),
			})
			return diags
		}
	}

	if d.HasChange("columns") {
		if err := alterColumn(db, d); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Error update the table name ",
				Detail:   "Unable to update the name of the table" + err.Error(),
			})
			return diags
		}
	}

	return resourcePostgreSqlTableRead(ctx, d, m)
}

func alterTableName(db *DBConnection, d *schema.ResourceData) error {
	schemaName := d.Get("schema").(string)
	var oldTableName, newTableName = d.GetChange("table")
	completeOldTableName := schemaName + "." + pq.QuoteIdentifier(oldTableName.(string))
	sql := fmt.Sprintf("ALTER TABLE %s RENAME TO %s", completeOldTableName, pq.QuoteIdentifier(newTableName.(string)))
	if _, err := db.Exec(sql); err != nil {
		d.Set("table", oldTableName)
		return fmt.Errorf("Error update table name %q to %q: %w ", oldTableName, newTableName, err)
	}
	return nil
}

func alterColumn(db *DBConnection, d *schema.ResourceData) error {
	schemaName := d.Get("schema").(string)
	tableName := d.Get("table").(string)
	completeOldTableName := schemaName + "." + pq.QuoteIdentifier(tableName)

	var oldColumns, newColumns = d.GetChange("columns")
	var err error

	if len(oldColumns.([]interface{})) <= len(newColumns.([]interface{})) {
		newSlice := newColumns.([]interface{})[len(oldColumns.([]interface{})):]
		for _, newColumn := range newSlice {
			newCol := newColumn.(map[string]interface{})
			sql := fmt.Sprintf("ALTER TABLE %s  ADD COLUMN %q %s", completeOldTableName, newCol["name"], newCol["type"])
			err = executeQuery(db, sql)
		}
	}

	if len(oldColumns.([]interface{})) > len(newColumns.([]interface{})) {
		newSlice := oldColumns.([]interface{})[len(newColumns.([]interface{})):]
		for _, newColumn := range newSlice {
			newCol := newColumn.(map[string]interface{})
			sql := fmt.Sprintf("ALTER TABLE %s  DROP COLUMN %q RESTRICT", completeOldTableName, newCol["name"])
			err = executeQuery(db, sql)
		}
	}

	for i, newColumn := range newColumns.([]interface{}) {
		newCol := newColumn.(map[string]interface{})
		if i <= len(oldColumns.([]interface{}))-1 {
			oldCol := oldColumns.([]interface{})[i].(map[string]interface{})
			if newCol["name"] != oldCol["name"] {
				sql := fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %q TO %q", completeOldTableName, oldCol["name"], newCol["name"])
				err = executeQuery(db, sql)
			}
			if newCol["type"] != oldCol["type"] {
				sql := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %q TYPE %s USING (%q::%s)", completeOldTableName, newCol["name"], newCol["type"], newCol["name"], newCol["type"])
				err = executeQuery(db, sql)
			}
		}
	}

	return err
}

func executeQuery(db *DBConnection, sql string) error {
	if _, err := db.Exec(sql); err != nil {
		return fmt.Errorf("Error running sql query  %q: %s", err, sql)
	}
	return nil
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
		return fmt.Errorf("Error creating table %q: %w", tableName, err)
	}

	var err error

	return err
}

func resourcePostgreSqlTableRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	db := m.(*DBConnection)
	return getTableColumnsDefinition(db, d)
}

func resourcePostgreSqlTableDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	d.SetId("")
	return diags

}

func getTableColumnsDefinition(db *DBConnection, d *schema.ResourceData) diag.Diagnostics {
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

	var name, columnType string
	for rows.Next() {
		if err := rows.Scan(&name, &columnType); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Could not scan tables columns",
				Detail:   "Unable to scan tables columns",
			})
			return diags
		}
		column := make(map[string]interface{})
		column["name"] = name
		column["type"] = columnType
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
