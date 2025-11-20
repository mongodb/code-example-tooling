package main

import (
	"common"
	"strings"
)

// GetProductSubProduct returns the product taxonomy for a given page in a project, where the project corresponds to a
// collection in our code example database. It uses predefined mappings from the `common` package to determine the product
// and sub-product, if any, based on the project name and page URL.
// NOTE: If the project is `cloud-docs` and the page ID contains a subdirectory string that corresponds to a mapped Atlas
// sub-product, the function returns that string.
func GetProductSubProduct(project string, page string) (string, string) {
	var productInfo common.ProductInfo

	// If the project is `cloud-docs`, the subdirectory of the docs may correspond with one of the subproductdir strings.
	// Each of them represents a different sub-product of Atlas. If the string is present in the page ID, return the
	// corresponding product info.
	if project == "cloud-docs" {
		subProductStringKeys := common.SubProductDirs
		for _, dir := range subProductStringKeys {
			if strings.Contains(page, dir) {
				productInfo = common.GetProductInfo(dir)
			}
		}
		// If the project is cloud-docs and we didn't find a sub-product in the page ID, just return the product info
		// for the project itself.
		if productInfo.ProductName == "" {
			productInfo = common.GetProductInfo(project)
		}
	} else {
		// Otherwise, just get the product/sub-product info defined in the common package
		productInfo = common.GetProductInfo(project)
	}
	return productInfo.ProductName, productInfo.SubProduct
}
