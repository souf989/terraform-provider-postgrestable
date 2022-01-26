package postgresql

import "fmt"

func contains(s []interface{}, str interface{}) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func getRightDiffColumns(oldColumns interface{}, newColumns interface{}) []map[string]interface{} {
	var diffColumns []map[string]interface{}
	var keyContainer []interface{}
	for _, value := range oldColumns.([]interface{}) {
		keyContainer = append(keyContainer, value.(map[string]interface{})["name"])
	}
	for _, value := range newColumns.([]interface{}) {
		newValue := value.(map[string]interface{})
		if !contains(keyContainer, newValue["name"]) {
			diffColumns = append(diffColumns, newValue)
		}
	}

	return diffColumns
}

func checkIfDuplicateColumns(columns interface{}) error {

	var container []interface{}
	for _, newColumn := range columns.([]interface{}) {
		newCol := newColumn.(map[string]interface{})
		key := newCol["name"]
		if contains(container, key) {
			return fmt.Errorf(" : there is a duplicate key : %q", newColumn)
		}
		container = append(container, key)
	}
	return nil
}

func executeQuery(db *DBConnection, sql string) error {
	if _, err := db.Exec(sql); err != nil {
		return fmt.Errorf("Error running sql query  %q: %s", err, sql)
	}
	return nil
}
