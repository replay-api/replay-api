package common_test

import (
	"testing"
	"time"

	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
)

func TestValidateSearchParameters(t *testing.T) {
	invalidDuration := time.Duration(-1)
	invalidDuration2 := time.Duration(0)

	validDuration := time.Duration(1)
	validDuration2 := time.Duration(2)

	// timeNow := time.Now()

	tests := []struct {
		name              string
		searchParams      []common.SearchAggregation
		queryableFields   map[string]bool
		expectedError     string
		maxRecursiveDepth int
	}{
		{
			name: "Valid Single Value Param",
			searchParams: []common.SearchAggregation{
				{
					Params: []common.SearchParameter{
						{
							ValueParams: []common.SearchableValue{
								{Field: "GameID", Values: []interface{}{"123"}},
							},
						},
					},
				},
			},
			queryableFields: map[string]bool{"GameID": true},
			expectedError:   "",
		},
		{
			name: "Invalid Single Value Param",
			searchParams: []common.SearchAggregation{
				{
					Params: []common.SearchParameter{
						{
							ValueParams: []common.SearchableValue{
								{Field: "InvalidField", Values: []interface{}{"123"}},
							},
						},
					},
				},
			},
			queryableFields: map[string]bool{"GameID": true},
			expectedError:   "filtering on ValueParams field 'InvalidField' is not permitted",
		},
		{
			name: "Valid Wildcard Value Param",
			searchParams: []common.SearchAggregation{
				{
					Params: []common.SearchParameter{
						{
							ValueParams: []common.SearchableValue{
								{Field: "Header.*", Values: []interface{}{"123"}},
							},
						},
					},
				},
			},
			queryableFields: map[string]bool{"Header.Filestamp": true},
			expectedError:   "",
		},
		{
			name: "Invalid Wildcard Value Param",
			searchParams: []common.SearchAggregation{
				{
					Params: []common.SearchParameter{
						{
							ValueParams: []common.SearchableValue{
								{Field: "InvalidPrefix.*", Values: []interface{}{"123"}},
							},
						},
					},
				},
			},
			queryableFields: map[string]bool{"Header.Filestamp": true},
			expectedError:   "filtering on ValueParams fields matching 'InvalidPrefix.*' is not permitted",
		},
		{
			name: "Valid Date Param",
			searchParams: []common.SearchAggregation{
				{
					Params: []common.SearchParameter{
						{
							DateParams: []common.SearchableDateRange{
								{Field: "Timestamp", Min: &time.Time{}, Max: &time.Time{}},
							},
						},
					},
				},
			},
			queryableFields: map[string]bool{"Timestamp": true},
			expectedError:   "",
		},
		{
			name: "Valid Duration Param",
			searchParams: []common.SearchAggregation{
				{
					Params: []common.SearchParameter{
						{
							DurationParams: []common.SearchableDurationRange{
								{Field: "Duration", Min: &validDuration, Max: &validDuration2},
							},
						},
					},
				},
			},
			queryableFields: map[string]bool{"Duration": true},
			expectedError:   "",
		},
		{
			name: "Invalid Duration Param",
			searchParams: []common.SearchAggregation{
				{
					Params: []common.SearchParameter{
						{
							DurationParams: []common.SearchableDurationRange{
								{Field: "InvalidField", Min: &invalidDuration, Max: &invalidDuration2},
							},
						},
					},
				},
			},
			queryableFields: map[string]bool{"Duration": true},
			expectedError:   "filtering on DurationParams field 'InvalidField' is not permitted",
		},
		{
			name: "Valid Recursive Depth",
			searchParams: []common.SearchAggregation{
				{
					Params: []common.SearchParameter{
						{
							ValueParams: []common.SearchableValue{
								{Field: "Header.Filestamp", Values: []interface{}{"HLTV"}},
							},
						},
					},
				},
			},
			queryableFields:   map[string]bool{"Header.Filestamp": true},
			expectedError:     "",
			maxRecursiveDepth: 1,
		},
		{
			name: "Invalid Recursive Depth",
			searchParams: []common.SearchAggregation{
				{
					Params: []common.SearchParameter{
						{
							ValueParams: []common.SearchableValue{
								{Field: "Header.Filestamp", Values: []interface{}{"123"}},
							},
						},
					},
				},
			},
			queryableFields:   map[string]bool{"Header.Filestamp": false},
			expectedError:     "filtering on ValueParams field 'Header.Filestamp' is not permitted",
			maxRecursiveDepth: 0,
		},
		{
			name: "Invalid Wildcard in Nested Field (Date)",
			searchParams: []common.SearchAggregation{
				{
					Params: []common.SearchParameter{
						{
							DateParams: []common.SearchableDateRange{
								{Field: "Header.InvalidSubField.*", Min: &time.Time{}, Max: &time.Time{}},
							},
						},
					},
				},
			},
			queryableFields: map[string]bool{"Header.Filestamp": true},
			expectedError:   "filtering on DateParams fields matching 'Header.InvalidSubField.*' is not permitted",
		},
		{
			name: "Invalid Wildcard in Nested Field (Duration)",
			searchParams: []common.SearchAggregation{
				{
					Params: []common.SearchParameter{
						{
							DurationParams: []common.SearchableDurationRange{
								{Field: "Header.InvalidSubField.*", Min: &validDuration, Max: &validDuration2},
							},
						},
					},
				},
			},
			queryableFields: map[string]bool{"Header.Filestamp": true},
			expectedError:   "filtering on DurationParams fields matching 'Header.InvalidSubField.*' is not permitted",
		},
		{
			name: "Disallowed Field (ValueParam) with Wildcard Allowed",
			searchParams: []common.SearchAggregation{
				{
					Params: []common.SearchParameter{
						{
							ValueParams: []common.SearchableValue{
								{Field: "Header.Filestamp", Values: []interface{}{"123"}},
							},
						},
					},
				},
			},
			queryableFields: map[string]bool{"Header.*": true, "Header.Filestamp": false}, // Filestamp disallowed specifically
			expectedError:   "filtering on ValueParams field 'Header.Filestamp' is not permitted",
		},

		// // Invalid Date Range - Start After End
		// {
		// 	name: "Invalid Date Param - Start After End",
		// 	searchParams: []common.SearchAggregation{
		// 		{
		// 			Params: []common.SearchParameter{
		// 				{
		// 					DateParams: []common.SearchableDateRange{
		// 						{Field: "Timestamp", Min: func() *time.Time {
		// 							t := timeNow.Add(time.Microsecond * 1)
		// 							return &t
		// 						}(), Max: &timeNow,
		// 						},
		// 					},
		// 				},
		// 			},
		// 		},
		// 		queryableFields: map[string]bool{"Timestamp": true},
		// 		expectedError:   "",
		// 	},

		// 	// Invalid Duration Range - Start After End
		// 	{
		// 		name: "Invalid Duration Param - Start After End",
		// 		searchParams: []common.SearchAggregation{
		// 			{
		// 				Params: []common.SearchParameter{
		// 					{
		// 						DurationParams: []common.SearchableDurationRange{
		// 							{Field: "Duration", Min: &validDuration2, Max: &validDuration},
		// 						},
		// 					},
		// 				},
		// 			},
		// 		},
		// 		queryableFields: map[string]bool{"Duration": true},
		// 		expectedError:   "",
		// 	},

		// 	{
		// 		name: "Missing 'Min' or 'Max' in Date Range",
		// 		searchParams: []common.SearchAggregation{
		// 			{
		// 				Params: []common.SearchParameter{
		// 					{
		// 						DateParams: []common.SearchableDateRange{
		// 							{Field: "Timestamp", Min: nil, Max: timeNow}, // Valid
		// 							{Field: "Timestamp", Min: timeNow, Max: nil}, // Valid
		// 						},
		// 					},
		// 				},
		// 			},
		// 		},
		// 		queryableFields: map[string]bool{"Timestamp": true},
		// 		expectedError:   "",
		// 	},

		// 	// Missing 'Min' or 'Max' in Duration Range - Should be valid
		// 	{
		// 		name: "Missing 'Min' or 'Max' in Duration Range",
		// 		searchParams: []common.SearchAggregation{
		// 			{
		// 				Params: []common.SearchParameter{
		// 					{
		// 						DurationParams: []common.SearchableDurationRange{
		// 							{Field: "Duration", Min: nil, Max: &validDuration},
		// 							{Field: "Duration", Min: &validDuration, Max: nil},
		// 						},
		// 					},
		// 				},
		// 			},
		// 		},
		// 		queryableFields: map[string]bool{"Duration": true},
		// 		expectedError:   "",
		// 	},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := common.ValidateSearchParameters(tt.searchParams, tt.queryableFields)
			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error '%s', but got no error", tt.expectedError)
				} else if err.Error() != tt.expectedError {
					t.Errorf("Expected error '%s', but got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got '%s'", err.Error())
				}
			}
		})
	}
}

func TestValidateResultOptions(t *testing.T) {
	tests := []struct {
		name           string
		resultOptions  common.SearchResultOptions
		readableFields map[string]bool
		expectedError  string
		maxPageSize    uint
	}{
		{
			name:           "Valid Result Options",
			resultOptions:  common.SearchResultOptions{Skip: 0, Limit: 5},
			readableFields: map[string]bool{"GameID": true},
			expectedError:  "",
		},
		{
			name:           "Invalid Limit",
			resultOptions:  common.SearchResultOptions{Skip: 0, Limit: 0},
			readableFields: map[string]bool{"GameID": true},
			expectedError:  "limit must be a positive integer",
		},
		{
			name:           "Invalid Pick Field with Wildcard Allowed",
			resultOptions:  common.SearchResultOptions{PickFields: []string{"Header.NonExistentField"}},
			readableFields: map[string]bool{"Header.*": true},
			expectedError:  "returning field 'Header.NonExistentField' is not permitted (1)",
		},
		{
			name:           "Invalid Omit Field with Wildcard Allowed",
			resultOptions:  common.SearchResultOptions{OmitFields: []string{"Header.NonExistentField"}},
			readableFields: map[string]bool{"Header.*": true},
			expectedError:  "omitting field 'Header.NonExistentField' is not permitted",
		},
		{
			name:           "Disallowed Pick Field Even with Wildcard",
			resultOptions:  common.SearchResultOptions{PickFields: []string{"Header.FileStamp"}},
			readableFields: map[string]bool{"Header.*": true, "Header.FileStamp": false},
			expectedError:  "returning field 'Header.FileStamp' is strictly forbidden (2)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := common.ValidateResultOptions(tt.resultOptions, tt.readableFields)
			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error '%s', but got no error", tt.expectedError)
				} else if err.Error() != tt.expectedError {
					t.Errorf("Expected error '%s', but got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got '%s'", err.Error())
				}
			}
		})
	}
}
