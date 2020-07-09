package notifier

import (
	"fmt"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	v1 "k8s.io/api/core/v1"
)

type severity string

// DANGER, WARNING AND GOOD are various severity levels for notifications.
const (
	DANGER  severity = "#FF0000"
	WARNING severity = "#FFFF00"
	GOOD    severity = "#006400"
)

//Notifier is a notification engine
type Notifier struct {
	messageClient MessageClient
}

// MessageClient is a Messaging interface
type MessageClient interface {
	push(severity severity, title, details string) error
}

//NewNotifier creates a new notifier client
func NewNotifier(client MessageClient) Notifier {
	return Notifier{messageClient: client}
}

//AnnotateNode sends notification about annotation added to node
func (n Notifier) AnnotateNode(node v1.Node) error {
	severity := GOOD
	title := "ANNOTATE"
	details := fmt.Sprintf("Node: %s\nCreation Time: %s\nExpiryTime: %s", node.Name, node.CreationTimestamp, node.Annotations[config.SpotterExpiryTimeAnnotation])

	err := n.messageClient.push(severity, title, details)
	return err
}

// DrainNode sends notification about draining a node
func (n Notifier) DrainNode(node v1.Node) error {

	severity := GOOD
	title := "DRAIN"
	details := fmt.Sprintf("Node: %s\nCreation Time: %s\nExpiryTime: %s", node.Name, node.CreationTimestamp, node.Annotations[config.SpotterExpiryTimeAnnotation])

	err := n.messageClient.push(severity, title, details)
	return err
}

//DrainNodeTimeout sends notification about node draining timing out
func (n Notifier) DrainNodeTimeout(node v1.Node) error {
	severity := DANGER
	title := "DRAIN TIMEOUT"
	details := fmt.Sprintf("Node: %s\nCreation Time: %s\nExpiryTime: %s", node.Name, node.CreationTimestamp, node.Annotations[config.SpotterExpiryTimeAnnotation])

	err := n.messageClient.push(severity, title, details)
	return err
}

//DeleteNode sends notification about deleting node
func (n Notifier) DeleteNode(node v1.Node) error {
	severity := GOOD
	title := "DELETE NODE"
	details := fmt.Sprintf("Node: %s\nCreation Time: %s\nExpiryTime: %s", node.Name, node.CreationTimestamp, node.Annotations[config.SpotterExpiryTimeAnnotation])

	err := n.messageClient.push(severity, title, details)
	return err
}

//DeletingInstance sends notification about gcp instance deletion.
func (n Notifier) DeleteInstance(node v1.Node) error {
	severity := GOOD
	title := "DELETE INSTANCE"
	details := fmt.Sprintf("Node: %s\nCreation Time: %s\nExpiryTime: %s", node.Name, node.CreationTimestamp, node.Annotations[config.SpotterExpiryTimeAnnotation])

	err := n.messageClient.push(severity, title, details)
	return err
}

//FailedToAnnotateNode sends notification about failing to annotate node
func (n Notifier) FailedToAnnotateNode(node v1.Node, err error) error {
	severity := DANGER
	title := "ERROR ANNOTATE"
	details := fmt.Sprintf("Node: %s\nCreation Time: %s\nExpiryTime: %s\nError: %s", node.Name, node.CreationTimestamp, node.Annotations[config.SpotterExpiryTimeAnnotation], err.Error())

	err = n.messageClient.push(severity, title, details)
	return err
}

//FailedToDrainNode sends notification about failing to drain node
func (n Notifier) FailedToDrainNode(node v1.Node, err error) error {
	severity := DANGER
	title := "ERROR DRAIN NODE"
	details := fmt.Sprintf("Node: %s\nCreation Time: %s\nExpiryTime: %s\nError: %s", node.Name, node.CreationTimestamp, node.Annotations[config.SpotterExpiryTimeAnnotation], err.Error())

	err = n.messageClient.push(severity, title, details)
	return err
}

//FailedToDeleteNode sends notification about failing to delete node
func (n Notifier) FailedToDeleteNode(node v1.Node, err error) error {
	severity := DANGER
	title := "ERROR DELETE NODE"
	details := fmt.Sprintf("Node: %s\nCreation Time: %s\nExpiryTime: %s\nError: %s", node.Name, node.CreationTimestamp, node.Annotations[config.SpotterExpiryTimeAnnotation], err.Error())

	err = n.messageClient.push(severity, title, details)
	return err
}

//FailedToDeleteInstance sends notification about failing to delete gcp instance
func (n Notifier) FailedToDeleteInstance(node v1.Node, err error) error {
	severity := DANGER
	title := "ERROR DELETE INSTANCE"
	details := fmt.Sprintf("Node: %s\nCreation Time: %s\nExpiryTime: %s\nError: %s", node.Name, node.CreationTimestamp, node.Annotations[config.SpotterExpiryTimeAnnotation], err.Error())

	err = n.messageClient.push(severity, title, details)
	return err
}
