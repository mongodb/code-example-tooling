package add_code_examples

func GetCodeExampleForTesting(category string, languageCategory string) string {
	driverUsageExample := `const bson_t *doc;
bson_t *pipeline = BCON_NEW ("pipeline",
    "[", 
    "{", "$match", "{", "cuisine", BCON_UTF8 ("Bakery"), "}", "}",
    "{", "$group", "{", 
        "_id", BCON_UTF8 ("$borough"), "count", "{", "$sum", BCON_INT32 (1), "}", "}",
    "}",
    "]");

mongoc_cursor_t *results =
    mongoc_collection_aggregate (collection, MONGOC_QUERY_NONE, pipeline, NULL, NULL);

bson_error_t error;
if (mongoc_cursor_error (results, &error))
{
    fprintf (stderr, "Aggregate failed: %s\n", error.message);
} else {
    while (mongoc_cursor_next (results, &doc)) {
        char *str = bson_as_canonical_extended_json (doc, NULL);
        printf ("%s\n", str);
        bson_free (str);
    }
}

bson_destroy (pipeline);
mongoc_cursor_destroy (results);`

	driverSyntaxExample := `auto_encryption_opts = AutoEncryptionOpts(mongocryptd_spawn_args=['--idleShutdownTimeoutSecs=75'])`

	driverReturnExample := `{ "_id" : "Queens", "count" : { "$numberInt" : "204" } }
{ "_id" : "Staten Island", "count" : { "$numberInt" : "20" } }
{ "_id" : "Missing", "count" : { "$numberInt" : "2" } }
{ "_id" : "Bronx", "count" : { "$numberInt" : "71" } }
{ "_id" : "Brooklyn", "count" : { "$numberInt" : "173" } }
{ "_id" : "Manhattan", "count" : { "$numberInt" : "221" } }`

	jsonLikeUsageExample := `db.movies.aggregate([
  {
    "$search": {
      "near": {
        "path": "released",
        "origin": ISODate("2011-09-01T00:00:00.000+00:00"),
        "pivot": 7776000000
      }
    }
  },
  {
    $project: {
      "_id": 0,
      "title": 1,
      "released": 1
    }
  },
  { "$limit": 5 },
  {
    "$facet": {
      "docs": [],
      "meta": [
        {"$replaceWith": "$$SEARCH_META"},
        {"$limit": 1}
      ]
    }
  }
])`

	jsonLikeSyntaxExample := `{
  $search: {
    "index": "<index-name>",
    "<operator-name>"|"<collector-name>": {
      <operator-specification>|<collector-specification>
    },
    "highlight": {
      <highlight-options>
    },
    "concurrent": true | false,
    "count": {
      <count-options>
    },
    "searchAfter"|"searchBefore": "<encoded-token>",
    "scoreDetails": true| false,
    "sort": {
      <fields-to-sort>: 1 | -1
    },
    "returnStoredSource": true | false,
    "tracking": {
      <tracking-option>
    }
   }
}`

	jsonLikeReturnExample := `{
  "detail" : "Cannot find resource /api/atlas/v1.0/softwareComponents/version.",
  "error" : 404,
  "errorCode" : "RESOURCE_NOT_FOUND",
  "parameters" : [ "/api/atlas/v1.0/softwareComponents/version" ],
  "reason" : "Not Found"
}`

	jsonLikeConfigurationExample := `{
  "key1_string": "value1",
  "key2_number": 1234,
  "key3_array": ["a", "b" ],
  "key4_dict": { subkey1: "subkeyvalue" }
}`

	javaScriptUsageExample := `db.movies.aggregate([
  {
    "$search": {
      "text": {
        "query": "liberte",
        "path": { "value": "title", "multi": "frenchAnalyzer" }
      }
    }
  },
  {
    "$project": {
      "title": 1,
      "year": 1,
      "_id": 0
    }
  }
])`

	javaScriptSyntaxExample := `myColl.find({
   rating: 5
});`

	javaScriptReturnExample := `{
  "_id": 0,
  "title": 1,
}`

	javaScriptConfigurationExample := `{
  "mappings": {
    "dynamic": true,
    "fields": {
      "items": [
        {
          "dynamic": true,
          "type": "embeddedDocuments"
        },
        {
          "dynamic": true,
          "fields": {
            "tags": {
              "type": "token"
            }
          },
          "type": "document"
        }
      ],
      "purchaseMethod": {
        "type": "stringFacet"
      }
    }
  }
}`

	shellUsageExample := `db.cars.aggregate([
  {
    $search: {
      "text": {
        "query": "pour",
        "path": "subject.fr"
      }
    }
  },
  {
    $project: {
      "_id": 0,
      "subject.fr": 1
    }
  }
])`

	shellSyntaxExample := `atlas accessLogs list [options]`

	shellReturnExample := `MongoError: Remote error from mongot :: caused by :: Error connecting to localhost:28000.`

	shellConfigurationExample := `mongos --setParameter 'healthMonitoringIntensities={ values:[ { type:"dns", intensity: "critical"} ] }'`

	shellNonMongoDBCommandExample := `sudo systemctl start mongod`

	textUsageExample := `keyVault = db.getKeyVault()

keyVault.rewrapManyDataKey(
   {
      "<Your custom filter>"
   },
   {
      provider: "<KMS provider>",
      masterKey: {
         "<dataKeyOpts Key>" : "<dataKeyOpts Value>"
      }
   }
)`

	textSyntaxExample := `{
  "fname" : "<String>",
  "lname" : "<String>",
  "passportId" : "<String>",
  "bloodType" : "<String>",
  "medicalRecords" : [
    {<object>}
  ],
  "insurance" : {
    "policyNumber" : "<string>",
    "provider" : "<string>"
  }
}`

	textReturnExample := `<dbpath>/rollback/20f74796-d5ea-42f5-8c95-f79b39bad190/removed.2020-02-19T04-57-11.0.bson`

	textConfigurationExample := `# NOT FOR PRODUCTION USE. OpenSSL configuration file for testing.


[ req ]
default_bits = 4096
default_keyfile = myTestServerCertificateKey.pem    ## The default private key file name.
default_md = sha256
distinguished_name = req_dn
req_extensions = v3_req

[ v3_req ]
subjectKeyIdentifier  = hash
basicConstraints = CA:FALSE
keyUsage = critical, digitalSignature, keyEncipherment
nsComment = "OpenSSL Generated Certificate for TESTING only.  NOT FOR PRODUCTION USE."
extendedKeyUsage  = serverAuth, clientAuth
subjectAltName = @alt_names

[ alt_names ]
DNS.1 =         ##TODO: Enter the DNS names. The DNS names should match the server names.
DNS.2 =         ##TODO: Enter the DNS names. The DNS names should match the server names.
IP.1 =          ##TODO: Enter the IP address.
IP.2 =          ##TODO: Enter the IP address.

[ req_dn ]
countryName = Country Name (2 letter code)
countryName_default = TestServerCertificateCountry
countryName_min = 2
countryName_max = 2

stateOrProvinceName = State or Province Name (full name)
stateOrProvinceName_default = TestServerCertificateState
stateOrProvinceName_max = 64

localityName = Locality Name (eg, city)
localityName_default = TestServerCertificateLocality
localityName_max = 64

organizationName = Organization Name (eg, company)
organizationName_default = TestServerCertificateOrg
organizationName_max = 64

organizationalUnitName = Organizational Unit Name (eg, section)
organizationalUnitName_default = TestServerCertificateOrgUnit
organizationalUnitName_max = 64

commonName = Common Name (eg, YOUR name)
commonName_max = 64`

	textNonMongoDBCommandExample := `dn:CN=dba,CN=Users,dc=example,dc=com
dn:CN=admin,CN=Users,dc=example,dc=com`

	switch category {
	case UsageExample:

		switch languageCategory {
		case DriversMinusJs:
			return driverUsageExample
		case Shell:
			return shellUsageExample
		case JavaScript:
			return javaScriptUsageExample
		case JsonLike:
			return jsonLikeUsageExample
		case Text:
			return textUsageExample
		default:
			return ""
		}

	case SyntaxExample:

		switch languageCategory {
		case DriversMinusJs:
			return driverSyntaxExample
		case Shell:
			return shellSyntaxExample
		case JavaScript:
			return javaScriptSyntaxExample
		case JsonLike:
			return jsonLikeSyntaxExample
		case Text:
			return textSyntaxExample
		default:
			return ""
		}

	case ExampleReturnObject:

		switch languageCategory {
		case DriversMinusJs:
			return driverReturnExample
		case Shell:
			return shellReturnExample
		case JavaScript:
			return javaScriptReturnExample
		case JsonLike:
			return jsonLikeReturnExample
		case Text:
			return textReturnExample
		default:
			return ""
		}

	case ExampleConfigurationObject:

		switch languageCategory {
		case DriversMinusJs:
			return ""
		case Shell:
			return shellConfigurationExample
		case JavaScript:
			return javaScriptConfigurationExample
		case JsonLike:
			return jsonLikeConfigurationExample
		case Text:
			return textConfigurationExample
		default:
			return ""
		}

	case NonMongoCommand:

		switch languageCategory {
		case Shell:
			return shellNonMongoDBCommandExample
		case Text:
			return textNonMongoDBCommandExample
		default:
			return ""
		}

	default:
		return ""
	}
}
