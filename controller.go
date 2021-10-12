package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type reconciler struct {
	client client.Client
	cmd    string
}

var _ reconcile.Reconciler = &reconciler{}

func (r *reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx)

	d := &appsv1.Deployment{}
	err := r.client.Get(ctx, request.NamespacedName, d)
	if errors.IsNotFound(err) {
		log.Error(nil, "Could not find resource")
		return reconcile.Result{}, nil
	}

	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not fetch ReplicaSet: %+v", err)
	}

	log.Info("Reconciling resource", "resource", request.NamespacedName)

	tmpfile, err := ioutil.TempFile("", "kwatchexec")
	if err != nil {
		return reconcile.Result{}, err
	}

	enc := json.NewEncoder(tmpfile)
	enc.Encode(d)

	tmpfile.Close()

	// this should be idempotent
	cmd := exec.Command(r.cmd, tmpfile.Name())
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	log.Info("Result", "cmd", r.cmd, "result", err == nil)

	return reconcile.Result{}, err
}
