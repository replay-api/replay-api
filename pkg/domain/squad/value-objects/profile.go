package squad_value_objects

type Profile struct {
	Name    string                 `json:"name" bson:"name"`
	Details map[string]interface{} `json:"details" bson:"details"`
}

// i.e.
// social: {
//   twitter: string;
//   linkedin: string;
//   github?: string;
// };
