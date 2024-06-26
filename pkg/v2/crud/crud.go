package crud

import (
	"fmt"

	"github.com/imulab/go-scim/pkg/v2/crud/expr"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
)

// Add value to SCIM resource at the given SCIM path. If SCIM path is empty, value will be added
// to the root of the resource. The supplied value must be compatible with the target property attribute,
// otherwise error will be returned.
func Add(resource *prop.Resource, path string, value interface{}) error {
	if len(path) == 0 {
		return resource.Navigator().Add(value).Error()
	}

	head, err := expr.CompilePath(path)
	if err != nil {
		return err
	}

	return defaultTraverse(resource.RootProperty(), skipMainSchemaNamespace(resource, head), func(nav prop.Navigator) error {
		return nav.Add(value).Error()
	})
}

// Replace value in SCIM resource at the given SCIM path. If SCIM path is empty, the root of the resource
// will be replaced. The supplied value must be compatible with the target property attribute, otherwise
// error will be returned.
func Replace(resource *prop.Resource, path string, value interface{}) error {
	if len(path) == 0 {
		return resource.Navigator().Replace(value).Error()
	}

	head, err := expr.CompilePath(path)
	if err != nil {
		return err
	}

	return defaultTraverse(resource.RootProperty(), skipMainSchemaNamespace(resource, head), func(nav prop.Navigator) error {
		return nav.Replace(value).Error()
	})
}

// Delete value from the SCIM resource at the specified SCIM path. The path cannot be empty.
// value can be nil, or []interface{} with each element a map[string]interface{} or
// map[string]interface{}
func Delete(resource *prop.Resource, path string, value interface{}) error {
	if len(path) == 0 {
		return fmt.Errorf("%w: path must be specified for delete operation", spec.ErrInvalidPath)
	}

	head, err := expr.CompilePath(path)
	if err != nil {
		return err
	}
	var query *expr.Expression

	if value != nil {
		switch v := value.(type) {
		case map[string]interface{}:
			query, err = expr.FromValue(v)
		case []interface{}:
			query, err = expr.FromValueList(v)
		default:
			return fmt.Errorf("%w: values (%v) for a delete operation is of unsupported type %T",
				spec.ErrInvalidValue, value, value)
		}
	}
	if err != nil {
		return fmt.Errorf("%w: %v", spec.ErrInvalidValue, err)
	}
	head = head.Append(query)
	return defaultTraverse(
		resource.RootProperty(),
		skipMainSchemaNamespace(resource, head),
		func(nav prop.Navigator) error {
			return nav.Delete().Error()
		},
	)
}

func skipMainSchemaNamespace(resource *prop.Resource, query *expr.Expression) *expr.Expression {
	if query == nil {
		return nil
	}

	if query.IsPath() && query.Token() == resource.ResourceType().Schema().ID() {
		return query.Next()
	}

	return query
}
