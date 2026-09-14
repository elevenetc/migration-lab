package main

import "migration-lab/backend/internal/models"

type commentClasses struct {
	Predicted         models.PerformanceClass
	Observed          models.PerformanceClass
	PredictedComplete bool
	ObservedComplete  bool
}

// Summarize the most expensive measured statement without turning missing
// classifications into METADATA_ONLY. Partial coverage stays visible.
func commentPerformance(statements []models.StatementMeasurement) commentClasses {
	classes := commentClasses{PredictedComplete: len(statements) > 0, ObservedComplete: len(statements) > 0}
	for _, statement := range statements {
		if !knownCommentClass(statement.PredictedClass) {
			classes.PredictedComplete = false
		} else if classes.Predicted == "" || models.WorseThan(statement.PredictedClass, classes.Predicted) {
			classes.Predicted = statement.PredictedClass
		}
		if !knownCommentClass(statement.ObservedClass) {
			classes.ObservedComplete = false
		} else if classes.Observed == "" || models.WorseThan(statement.ObservedClass, classes.Observed) {
			classes.Observed = statement.ObservedClass
		}
	}
	return classes
}

func knownCommentClass(class models.PerformanceClass) bool {
	return class == models.MetadataOnly || class == models.DataScanning || class == models.TableRewrite
}

func renderCommentClass(class models.PerformanceClass, complete bool) string {
	if class == "" {
		return "Unavailable"
	}
	text := commentCode(string(class))
	if !complete {
		text += " (partial)"
	}
	return text
}
