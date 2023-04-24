package controllers

import (
	"html/template"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/zeebo/errs"
	"go.uber.org/zap"

	"todo/items"
)

var (
	// ErrItems is an internal error type for items controller.
	ErrItems = errs.Class("items controller error")
)

// ItemsTemplates holds all items related templates.
type ItemsTemplates struct {
	List   *template.Template
	Create *template.Template
	Update *template.Template
}

// Items is a mvc controller that handles all items related views.
type Items struct {
	log *zap.Logger

	items *items.Service

	templates ItemsTemplates
}

// NewItems is constructor for Items.
func NewItems(log *zap.Logger, items *items.Service, templates ItemsTemplates) *Items {
	return &Items{
		log:       log,
		items:     items,
		templates: templates,
	}
}

// Create is an endpoint that creates item.
func (controller *Items) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var err error
	params := mux.Vars(r)

	id, err := uuid.Parse(params["userId"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if err = controller.templates.Create.Execute(w, id); err != nil {
			controller.log.Error("could not parse template:" + ErrItems.Wrap(err).Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		if err = r.ParseForm(); err != nil {
			http.Error(w, "could not parse form", http.StatusBadRequest)
			return
		}

		name := r.FormValue("name")
		if name == "" {
			http.Error(w, "empty name field", http.StatusBadRequest)
			return
		}

		description := r.FormValue("description")
		if description == "" {
			http.Error(w, "empty name field", http.StatusBadRequest)
			return
		}

		if err = controller.items.Create(ctx, id, name, description); err != nil {
			controller.log.Error("could not update item:" + ErrItems.Wrap(err).Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		Redirect(w, r, "/"+id.String()+"/items", http.MethodGet)
	}
}

// List is an endpoint that returns all users items.
func (controller *Items) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)

	id, err := uuid.Parse(params["userId"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	allItems, err := controller.items.List(ctx, id)
	if err != nil {
		controller.log.Error("could not get items:" + ErrItems.Wrap(err).Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fields := struct {
		Items  []items.Item
		UserID uuid.UUID
	}{
		Items:  allItems,
		UserID: id,
	}

	if err = controller.templates.List.Execute(w, fields); err != nil {
		controller.log.Error("could not parse template:" + ErrItems.Wrap(err).Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Update is an endpoint that updates item.
func (controller *Items) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)

	userID, err := uuid.Parse(params["userId"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	request := struct {
		UserID uuid.UUID
		ItemID uuid.UUID
	}{
		UserID: userID,
		ItemID: id,
	}

	switch r.Method {
	case http.MethodGet:
		if err = controller.templates.Update.Execute(w, request); err != nil {
			controller.log.Error("could not parse template:" + ErrItems.Wrap(err).Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		if err = r.ParseForm(); err != nil {
			http.Error(w, "could not parse form", http.StatusBadRequest)
			return
		}

		name := r.FormValue("name")
		if name == "" {
			http.Error(w, "emp", http.StatusBadRequest)
			return
		}

		description := r.FormValue("description")
		if description == "" {
			http.Error(w, "emp", http.StatusBadRequest)
			return
		}

		if err = controller.items.Update(ctx, id, name, description); err != nil {
			controller.log.Error("could not create item:" + ErrItems.Wrap(err).Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		Redirect(w, r, "/"+userID.String()+"/items", http.MethodGet)
	}
}

// UpdateStatus is an endpoint that updates status users item.
func (controller *Items) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)

	userID, err := uuid.Parse(params["userId"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = controller.items.UpdateStatus(ctx, id); err != nil {
		controller.log.Error("could not update status of item:" + ErrItems.Wrap(err).Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	Redirect(w, r, "/"+userID.String()+"/items", http.MethodGet)
}

// Delete is an endpoint that delete users item.
func (controller *Items) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)

	userID, err := uuid.Parse(params["userId"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = controller.items.Delete(ctx, id); err != nil {
		controller.log.Error("could not delete item:" + ErrItems.Wrap(err).Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	Redirect(w, r, "/"+userID.String()+"/items", http.MethodGet)
}
