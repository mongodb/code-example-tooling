package common

const (
	CollectionIsProduct    = "collection"
	CollectionIsSubProduct = "collectionIsSubProduct"
	DirIsSubProduct        = "dirIsSubProduct"
	FocusArea              = "focusArea"
)

type ProductInfo struct {
	ProductName string
	ProductType string
	SubProduct  string
}

// GetProductInfo returns relevant product mapping for a given collection or page
func GetProductInfo(projectOrSubdir string) ProductInfo {
	if productInfo, found := productInfoMap[projectOrSubdir]; found {
		return productInfo
	}
	// Return a default or zero-value struct if the key wasn't found
	return ProductInfo{}
}

var SubProductDirs = []string{
	DataFederationDir,
	OnlineArchiveDir,
	StreamProcessingDir,
	SearchDir,
	TriggersDir,
	VectorSearchDir,
}

var productInfoMap = map[string]ProductInfo{
	"atlas-cli": {
		ProductName: Atlas,
		ProductType: CollectionIsSubProduct,
		SubProduct:  AtlasCLI,
	},
	"atlas-architecture": {
		ProductName: AtlasArchitecture,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"atlas-operator": {
		ProductName: Atlas,
		ProductType: CollectionIsSubProduct,
		SubProduct:  AtlasOperator,
	},
	"atlas-search": {
		ProductName: Atlas,
		ProductType: DirIsSubProduct,
		SubProduct:  Search,
	},
	"atlas-stream-processing": {
		ProductName: Atlas,
		ProductType: DirIsSubProduct,
		SubProduct:  StreamProcessing,
	},
	"atlas-vector-search": {
		ProductName: Atlas,
		ProductType: DirIsSubProduct,
		SubProduct:  VectorSearch,
	},
	"bi-connector": {
		ProductName: BIConnector,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"c": {
		ProductName: Drivers,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"charts": {
		ProductName: Atlas,
		ProductType: CollectionIsSubProduct,
		SubProduct:  Charts,
	},
	"cloud-docs": {
		ProductName: Atlas,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"cloud-manager": {
		ProductName: CloudManager,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"cloudgov": {
		ProductName: Atlas,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"compass": {
		ProductName: Compass,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"cpp-driver": {
		ProductName: Drivers,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"csharp": {
		ProductName: Drivers,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"database-tools": {
		ProductName: DBTools,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"data-federation": {
		ProductName: Atlas,
		ProductType: DirIsSubProduct,
		SubProduct:  DataFederation,
	},
	"django": {
		ProductName: Django,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"docs": {
		ProductName: Server,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"docs-k8s-operator": {
		ProductName: Atlas,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"docs-relational-migrator": {
		ProductName: RelationalMigrator,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"entity-framework": {
		ProductName: EFCoreProvider,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"golang": {
		ProductName: Drivers,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"java": {
		ProductName: Drivers,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"java-rs": {
		ProductName: Drivers,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"kafka-connector": {
		ProductName: KafkaConnector,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"kotlin": {
		ProductName: Drivers,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"kotlin-sync": {
		ProductName: Drivers,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"laravel": {
		ProductName: Drivers,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"mck": {
		ProductName: EnterpriseKubernetesOperator,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"mcp-server": {
		ProductName: MCPServer,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"mongoid": {
		ProductName: Drivers,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"mongodb-shell": {
		ProductName: Mongosh,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"mongocli": {
		ProductName: MDBCLI,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"mongosync": {
		ProductName: Mongosync,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"node": {
		ProductName: Drivers,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"online-archive": {
		ProductName: Atlas,
		ProductType: DirIsSubProduct,
		SubProduct:  OnlineArchive,
	},
	"ops-manager": {
		ProductName: OpsManager,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"php-library": { // DOCSP-51020 to add to taxonomy/programmatic tagging
		ProductName: Drivers,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"pymongo": {
		ProductName: Drivers,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"pymongo-arrow": {
		ProductName: Drivers,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"ruby-driver": {
		ProductName: Drivers,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"rust": {
		ProductName: Drivers,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"scala": {
		ProductName: Drivers,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"spark-connector": {
		ProductName: SparkConnector,
		ProductType: CollectionIsProduct,
		SubProduct:  "",
	},
	"terraform": {
		ProductName: Atlas,
		ProductType: DirIsSubProduct,
		SubProduct:  Terraform,
	},
	"triggers": {
		ProductName: Atlas,
		ProductType: DirIsSubProduct,
		SubProduct:  Triggers,
	},
}
