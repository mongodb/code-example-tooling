package main

import "testing"

func TestGetProductSubProduct(t *testing.T) {
	type args struct {
		project string
		page    string
	}
	tests := []struct {
		name           string
		args           args
		wantProduct    string
		wantSubProduct string
	}{
		{"Should correctly set product no sub-product", args{project: "docs", page: "https://mongodb.com/docs/manual/administration/deploy-manage-self-managed-sharded-clusters"}, "Server", ""},
		{"Should correctly set product and sub-product by collection", args{project: "charts", page: "https://mongodb.com/docs/charts/add-lookup-field"}, "Atlas", "Charts"},
		{"Should correctly set product and sub-product by dir", args{project: "cloud-docs", page: "https://www.mongodb.com/docs/atlas/atlas-search/aggregation-stages/searchMeta"}, "Atlas", "Search"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotProduct, gotSubProduct := GetProductSubProduct(tt.args.project, tt.args.page)
			if gotProduct != tt.wantProduct {
				t.Errorf("For product got = %v, want %v", gotProduct, tt.wantProduct)
			}
			if gotSubProduct != tt.wantSubProduct {
				t.Errorf("For sub-product got = %v, want %v", gotSubProduct, tt.wantSubProduct)
			}
		})
	}
}
