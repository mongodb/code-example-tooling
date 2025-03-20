package data

import (
	"gdcd/types"
	"time"
)

func GetUnchangedNodes() (types.CodeNode, types.ASTNode) {
	codeText := `
#include <mongoc/mongoc.h>
#include <bson/bson.h>

int main(void) {

  mongoc_uri_t* uri = NULL;
  mongoc_client_t *client = NULL;
  mongoc_database_t *database = NULL;
  bson_t *ping = NULL, reply = BSON_INITIALIZER;
  bson_error_t error;

  mongoc_init();

  // Start example code here

  // End example code here

  database = mongoc_client_get_database (client, "admin");
    
  ping = BCON_NEW ("ping", BCON_INT32 (1));
   
  if (!mongoc_client_command_simple (client, "admin", ping, NULL, &reply, &error)) {
    fprintf (stderr, "%s\n", error.message);
    goto cleanup;
  }
  printf ("Pinged your deployment. You successfully connected to MongoDB!\n");

  cleanup:
    bson_destroy (&reply);
    bson_destroy (ping);
    mongoc_database_destroy (database);
    mongoc_client_destroy (client);
    mongoc_uri_destroy (uri);
    mongoc_cleanup ();
}`

	codeNode := types.CodeNode{
		Code:           codeText,
		Language:       "c",
		FileExtension:  ".c",
		Category:       "Usage example",
		SHA256Hash:     "422b3486866301e76fe74d41441f075a7640091dd9b110d0783134a3f4b2c69f",
		LLMCategorized: true,
		DateAdded:      time.Now(),
		IsRemoved:      false,
	}

	astNode := types.ASTNode{
		Type:           "code",
		Position:       types.Position{Start: types.PositionLine{Line: 51}},
		Children:       nil,
		Value:          "#include <mongoc/mongoc.h>\n#include <bson/bson.h>\n\nint main(void) {\n\n  mongoc_uri_t* uri = NULL;\n  mongoc_client_t *client = NULL;\n  mongoc_database_t *database = NULL;\n  bson_t *ping = NULL, reply = BSON_INITIALIZER;\n  bson_error_t error;\n\n  mongoc_init();\n\n  // Start example code here\n\n  // End example code here\n\n  database = mongoc_client_get_database (client, \"admin\");\n    \n  ping = BCON_NEW (\"ping\", BCON_INT32 (1));\n   \n  if (!mongoc_client_command_simple (client, \"admin\", ping, NULL, &reply, &error)) {\n    fprintf (stderr, \"%s\\n\", error.message);\n    goto cleanup;\n  }\n  printf (\"Pinged your deployment. You successfully connected to MongoDB!\\n\");\n\n  cleanup:\n    bson_destroy (&reply);\n    bson_destroy (ping);\n    mongoc_database_destroy (database);\n    mongoc_client_destroy (client);\n    mongoc_uri_destroy (uri);\n    mongoc_cleanup ();\n}",
		Lang:           "c",
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
	return codeNode, astNode
}
