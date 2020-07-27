package store

import (
	"go.mongodb.org/mongo-driver/bson"
)

// AggregationMatch helps generate aggregation object for $match
func AggregationMatch(matchCondition bson.M) bson.D {
	match := bson.D{}
	for k, v := range matchCondition {
		match = append(match, bson.E{k, v})
	}

	return bson.D{
		bson.E{"$match", match},
	}
}

// AggregationProject helps generate aggregation object for $project
func AggregationProject(projectCondition bson.M) bson.D {
	project := bson.D{}
	for k, v := range projectCondition {
		project = append(project, bson.E{k, v})
	}

	return bson.D{
		bson.E{"$project", project},
	}
}

// AggregationUnwind helps generate aggregation object for $unwind
func AggregationUnwind(key string) bson.D {
	return bson.D{
		bson.E{"$unwind", key},
	}
}

// AggregationAddFields helps generate aggregation object for $addFields
func AggregationAddFields(fields bson.M) bson.D {
	return bson.D{
		bson.E{
			"$addFields", fields,
		},
	}
}

// AggregationGroup helps generate aggregation object for $group
func AggregationGroup(id interface{}, groupConditions bson.D) bson.D {
	group := bson.D{bson.E{"_id", id}}
	group = append(group, groupConditions...)

	return bson.D{
		bson.E{
			"$group", group,
		},
	}
}

// GeoNearOption is an option for AggregationGeoNear help function
type GeoNearOption struct {
	GeoKey             string
	DistanceKey        string
	DistanceMultiplier float64
}

// AggregationGroup helps generate aggregation object for $geoNear
func AggregationGeoNear(coordinates []float64, distance int, options ...GeoNearOption) bson.D {
	geoNear := bson.M{
		"near": bson.M{
			"type":        "Point",
			"coordinates": coordinates,
		},
		"maxDistance": distance,
		"spherical":   true,
	}

	if len(options) > 0 {
		option := options[0]
		if option.GeoKey != "" {
			geoNear["key"] = option.GeoKey
		}

		if option.DistanceKey != "" {
			geoNear["distanceField"] = option.DistanceKey
			if option.DistanceMultiplier != 0 {
				geoNear["distanceMultiplier"] = option.DistanceMultiplier
			}
		}
	}

	return bson.D{
		bson.E{"$geoNear", geoNear},
	}
}

// AggregationSort sorts the documents by specified key
func AggregationSort(key string, order int) bson.D {
	sort := bson.D{bson.E{key, order}}
	return bson.D{
		bson.E{
			"$sort", sort,
		},
	}
}

// AggregationLimit limits the number of documents in the pipeline
func AggregationLimit(limit int64) bson.D {
	return bson.D{{"$limit", limit}}
}
