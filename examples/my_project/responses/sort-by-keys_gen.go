// Code generated by github.com/AlwxSin/simplegen, DO NOT EDIT.
package responses

import (
	"examples/my_project/models"
)

func UserListSortByKeys(vs []*models.User, keys []int) []*models.User {
	res := make([]*models.User, len(keys))
	for i, key := range keys {
		var appendable *models.User
		for _, v := range vs {
			if key == v.ID {
				appendable = v
				break
			}
		}
		res[i] = appendable
	}
	return res
}

func UserListByEmailSortByKeys(vs []*models.User, keys []string) []*models.User {
	res := make([]*models.User, len(keys))
	for i, key := range keys {
		var appendable *models.User
		for _, v := range vs {
			if key == v.Email {
				appendable = v
				break
			}
		}
		res[i] = appendable
	}
	return res
}