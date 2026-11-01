// Package evaldata holds the portable fixture relevance judgments.
package evaldata

import _ "embed"

//go:embed queries.json
var Queries []byte
