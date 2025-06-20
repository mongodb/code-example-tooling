package data

import "gdcd/types"

func GetNewASTNodes(numNodes int) []types.ASTNode {
	node1 := types.ASTNode{
		Type:           "code",
		Position:       types.Position{Start: types.PositionLine{Line: 51}},
		Children:       nil,
		Value:          "var objectDiscriminatorConvention = BsonSerializer.LookupDiscriminatorConvention(typeof(object));\nvar objectSerializer = new ObjectSerializer(objectDiscriminatorConvention, GuidRepresentation.Standard);\nBsonSerializer.RegisterSerializer(objectSerializer);",
		Lang:           "csharp",
		Copyable:       true,
		Entries:        nil,
		EnumType:       "",
		ID:             "",
		Domain:         "",
		Name:           "",
		Argument:       nil,
		Options:        nil,
		EmphasizeLines: types.EmphasizeLines{14, 16},
	}

	node2 := types.ASTNode{
		Type:           "code",
		Position:       types.Position{Start: types.PositionLine{Line: 51}},
		Children:       nil,
		Value:          "Some Code Here\nAnd some more code\n\nint main(void) {\n\n  mongoc_uri_t* uri = NULL;\n  mongoc_client_t *client = NULL;\n  mongoc_database_t *database = NULL;\n  bson_t *ping = NULL, reply = BSON_INITIALIZER;\n  bson_error_t error;\n\n  mongoc_init();\n\n  // Start example code here\n\n  // End example code here\n\n  database = mongoc_client_get_database (client, \"admin\");\n    \n  ping = BCON_NEW (\"ping\", BCON_INT32 (1));\n   \n  if (!mongoc_client_command_simple (client, \"admin\", ping, NULL, &reply, &error)) {\n    fprintf (stderr, \"%s\\n\", error.message);\n    goto cleanup;\n  }\n  printf (\"Pinged your deployment. You successfully connected to MongoDB!\\n\");\n\n  cleanup:\n    bson_destroy (&reply);\n    bson_destroy (ping);\n    mongoc_database_destroy (database);\n    mongoc_client_destroy (client);\n    mongoc_uri_destroy (uri);\n    mongoc_cleanup ();\n}",
		Lang:           "go",
		Copyable:       false,
		Entries:        nil,
		EnumType:       "",
		ID:             "",
		Domain:         "",
		Name:           "",
		Argument:       nil,
		Options:        nil,
		EmphasizeLines: nil,
	}

	if numNodes == 1 {
		return []types.ASTNode{node1}
	} else {
		return []types.ASTNode{node1, node2}
	}
}
