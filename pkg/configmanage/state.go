package configmanage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"

	log "github.com/sirupsen/logrus"
)

// The name of the state file managed by the application
var stateFilePath string = "glueprint-state.json"

// CreateStateFileIfNotExists populates an empty state file when one
// does not exist
func CreateStateFileIfNotExists() {
	emptyFile, _ := json.MarshalIndent(map[string]string{}, "", " ")
	if _, err := os.Stat(stateFilePath); os.IsNotExist(err) {
		err = ioutil.WriteFile(stateFilePath, emptyFile, 0600)
		if err != nil {
			log.Errorf("Error writing to state file: %s", err)
		}
	}
}

// DeleteFromState removes a resource from the state file
func DeleteFromState(data map[string]ManagedResource) ([]map[string]ManagedResource, error) {
	// Read state
	existingState := ReadFromState()

	// Determine if the ManagedResource exists already
	// Reference for the struct containing a resource's specification
	var resource string
	for r := range data {
		resource = r
	}

	// Determine if the resource exists in state
	var inState bool = false
	for _, stateResource := range existingState {
		_, inState = stateResource[resource]
	}

	// Remove if it does, warn if it doesn't
	if inState {
		// Remove
		var newState []map[string]ManagedResource
		for i, entry := range existingState {
			for name := range entry {
				if name == resource {
					newState = append(existingState[0:i], existingState[i+1:]...)
				}
			}
		}
		updatedState, err := json.Marshal(newState)
		if err != nil {
			log.Errorf("Error formatting state data: %s", err)
		}
		err = ioutil.WriteFile(stateFilePath, updatedState, 0600)
		if err != nil {
			log.Errorf("Error writing to state file: %s", err)
		}
		return newState, nil
	} else {
		log.Warnf("Resource %s not found in state file", resource)
		return nil, fmt.Errorf("resource %s not found in state file", resource)
	}
}

// ReadFromState parses objects stored in the state file
func ReadFromState() []map[string]ManagedResource {
	if _, err := os.Stat(stateFilePath); os.IsNotExist(err) {
		log.Fatalf("Error reading from state file: %s", err)
	}
	data, err := os.ReadFile(stateFilePath)
	if err != nil {
		log.Errorf("Error opening state file: %s", err)
	}

	// Handle empty state
	if string(data) == "{}" {
		return []map[string]ManagedResource{}
	} else {
		// Handle populated state
		var allState []map[string]ManagedResource
		err := json.Unmarshal([]byte(string(data)), &allState)
		if err != nil {
			log.Errorf("Error reading from state file: %s", err)
		}
		return allState
	}
}

// ReadOneFromState parses a single object stored in the state file
func ReadOneFromState(resource string) map[string]ManagedResource {
	if _, err := os.Stat(stateFilePath); os.IsNotExist(err) {
		log.Warnf("Error reading from state file: %s - if this is the first time running the application, it will be created", err)
		return map[string]ManagedResource{}
	}
	data, err := os.ReadFile(stateFilePath)
	if err != nil {
		log.Errorf("Error opening state file: %s", err)
	}

	// Handle empty state
	if string(data) == "{}" {
		log.Errorf("No entry for %s found in state file", resource)
		return map[string]ManagedResource{}
	} else {
		// Handle populated state
		var allState []map[string]ManagedResource
		err := json.Unmarshal([]byte(string(data)), &allState)
		if err != nil {
			log.Errorf("Error reading from state file: %s", err)
		}
		for _, res := range allState {
			for resourceNameFromState := range res {
				if resourceNameFromState == resource {
					return res
				}
			}
		}
	}
	return map[string]ManagedResource{}
}

// WriteToState formats a resource to be written to state on creation,
// update, or removal
func WriteToState(data map[string]ManagedResource) {
	CreateStateFileIfNotExists()
	// Read state
	existingState := ReadFromState()

	// Determine if the ManagedResource exists already
	// Reference for the struct containing a resource's specification
	var resource string
	var resourceConfiguration ManagedResource
	for r, c := range data {
		resource = r
		resourceConfiguration = c
	}

	// Determine if the resource exists in state
	var stateResourceConfiguration ManagedResource
	var inState bool = false
	var ogMap map[string]ManagedResource

	for _, stateResource := range existingState {
		ogMap = stateResource
		stateResourceConfiguration, inState = stateResource[resource]
	}

	//
	if inState {
		log.Infof("Resource %s found in state file", resource)
		// Determine if requested configuration matches state configuration
		configurationMatches := reflect.DeepEqual(resourceConfiguration, stateResourceConfiguration)
		// Zero diff if matches, delete and re-add if not
		if configurationMatches {
			log.Infof("Resource %s is in sync, no changes to apply", resource)
		} else {
			log.Infof("Resource %s has changes, updating state file", resource)
			// Remove
			newState, err := DeleteFromState(ogMap)
			if err != nil {
				log.Errorf("Error removing resource: %s", err)
			}
			// Add
			finalState := append(newState, data)
			updatedState, err := json.Marshal(finalState)
			if err != nil {
				log.Errorf("Error formatting state data: %s", err)
			}
			err = ioutil.WriteFile(stateFilePath, updatedState, 0600)
			if err != nil {
				log.Errorf("Error writing to state file: %s", err)
			}
		}
	} else {
		log.Infof("Adding resource %s to state file...", resource)
		// It doesn't exist, so create it
		existingState = append(existingState, data)
		updatedState, err := json.Marshal(existingState)
		if err != nil {
			log.Errorf("Error formatting state data: %s", err)
		}
		err = ioutil.WriteFile(stateFilePath, updatedState, 0600)
		if err != nil {
			log.Errorf("Error writing to state file: %s", err)
		}
	}
}
