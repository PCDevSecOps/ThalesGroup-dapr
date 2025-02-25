/*
Copyright 2023 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package processor

import (
	"context"
	"fmt"

	contriblock "github.com/dapr/components-contrib/lock"
	compapi "github.com/dapr/dapr/pkg/apis/components/v1alpha1"
	complock "github.com/dapr/dapr/pkg/components/lock"
	diag "github.com/dapr/dapr/pkg/diagnostics"
	"github.com/dapr/dapr/pkg/runtime/compstore"
	rterrors "github.com/dapr/dapr/pkg/runtime/errors"
	"github.com/dapr/dapr/pkg/runtime/meta"
)

type lock struct {
	registry  *complock.Registry
	compStore *compstore.ComponentStore
	meta      *meta.Meta
}

func (l *lock) init(ctx context.Context, comp compapi.Component) error {
	// create the component
	fName := comp.LogName()
	store, err := l.registry.Create(comp.Spec.Type, comp.Spec.Version, fName)
	if err != nil {
		diag.DefaultMonitoring.ComponentInitFailed(comp.Spec.Type, "creation", comp.ObjectMeta.Name)
		return rterrors.NewInit(rterrors.CreateComponentFailure, fName, err)
	}

	if store == nil {
		return nil
	}

	// initialization
	baseMetadata := l.meta.ToBaseMetadata(comp)
	props := baseMetadata.Properties

	err = store.InitLockStore(ctx, contriblock.Metadata{Base: baseMetadata})
	if err != nil {
		diag.DefaultMonitoring.ComponentInitFailed(comp.Spec.Type, "init", comp.ObjectMeta.Name)
		return rterrors.NewInit(rterrors.InitComponentFailure, fName, err)
	}

	// save lock related configuration
	l.compStore.AddLock(comp.ObjectMeta.Name, store)
	err = complock.SaveLockConfiguration(comp.ObjectMeta.Name, props)
	if err != nil {
		diag.DefaultMonitoring.ComponentInitFailed(comp.Spec.Type, "init", comp.ObjectMeta.Name)
		wrapError := fmt.Errorf("failed to save lock keyprefix: %s", err)
		return rterrors.NewInit(rterrors.InitComponentFailure, fName, wrapError)
	}

	diag.DefaultMonitoring.ComponentInitialized(comp.Spec.Type)

	return nil
}
