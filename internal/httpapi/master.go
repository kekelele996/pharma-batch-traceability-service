package httpapi

import (
	"net/http"

	"pharma-batch-traceability-service/internal/customer"
	"pharma-batch-traceability-service/internal/drug"
	"pharma-batch-traceability-service/internal/manufacturer"
	"pharma-batch-traceability-service/internal/platform"
	"pharma-batch-traceability-service/internal/supplier"
	"pharma-batch-traceability-service/internal/warehouse"
)

func (s *Server) bulkSuspendCustomers(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := platform.DecodeJSON(r, &body); err != nil {
		platform.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	updated, err := s.customers.SuspendMany(body.IDs)
	if err != nil {
		platform.WriteJSON(w, http.StatusOK, map[string]any{"suspended": updated, "error": err.Error()})
		return
	}
	platform.WriteJSON(w, http.StatusOK, map[string]any{"suspended": updated})
}

func (s *Server) createDrug(w http.ResponseWriter, r *http.Request) {
	var d drug.Drug
	if err := platform.DecodeJSON(r, &d); err != nil {
		platform.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	created, err := s.drugs.Create(d)
	if err != nil {
		writeErr(w, err)
		return
	}
	s.metric.Inc("drugs.created", 1)
	platform.WriteJSON(w, http.StatusCreated, created)
}

func (s *Server) listDrugs(w http.ResponseWriter, r *http.Request) {
	if q := r.URL.Query().Get("q"); q != "" {
		platform.WriteJSON(w, http.StatusOK, s.drugs.Search(q))
		return
	}
	platform.WriteJSON(w, http.StatusOK, s.drugs.List())
}

func (s *Server) getDrug(w http.ResponseWriter, r *http.Request) {
	d, err := s.drugs.Get(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, d)
}

func (s *Server) updateDrug(w http.ResponseWriter, r *http.Request) {
	var d drug.Drug
	if err := platform.DecodeJSON(r, &d); err != nil {
		platform.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	updated, err := s.drugs.Update(pathID(r, "id"), d)
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, updated)
}

func (s *Server) createManufacturer(w http.ResponseWriter, r *http.Request) {
	var m manufacturer.Manufacturer
	if err := platform.DecodeJSON(r, &m); err != nil {
		platform.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	created, err := s.manufacturers.Create(m)
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusCreated, created)
}

func (s *Server) listManufacturers(w http.ResponseWriter, r *http.Request) {
	platform.WriteJSON(w, http.StatusOK, s.manufacturers.List())
}

func (s *Server) setManufacturerStatus(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}
	if err := platform.DecodeJSON(r, &body); err != nil {
		platform.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	updated, err := s.manufacturers.SetStatus(pathID(r, "id"), body.Status)
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, updated)
}

func (s *Server) createWarehouse(w http.ResponseWriter, r *http.Request) {
	var wh warehouse.Warehouse
	if err := platform.DecodeJSON(r, &wh); err != nil {
		platform.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	created, err := s.warehouses.Create(wh)
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusCreated, created)
}

func (s *Server) listWarehouses(w http.ResponseWriter, r *http.Request) {
	platform.WriteJSON(w, http.StatusOK, s.warehouses.List())
}

func (s *Server) setWarehouseStatus(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}
	if err := platform.DecodeJSON(r, &body); err != nil {
		platform.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	updated, err := s.warehouses.SetStatus(pathID(r, "id"), body.Status)
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, updated)
}

func (s *Server) createSupplier(w http.ResponseWriter, r *http.Request) {
	var v supplier.Supplier
	if err := platform.DecodeJSON(r, &v); err != nil {
		platform.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	created, err := s.suppliers.Create(v)
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusCreated, created)
}

func (s *Server) listSuppliers(w http.ResponseWriter, r *http.Request) {
	platform.WriteJSON(w, http.StatusOK, s.suppliers.List())
}

func (s *Server) createCustomer(w http.ResponseWriter, r *http.Request) {
	var v customer.Customer
	if err := platform.DecodeJSON(r, &v); err != nil {
		platform.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	created, err := s.customers.Create(v)
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusCreated, created)
}

func (s *Server) listCustomers(w http.ResponseWriter, r *http.Request) {
	platform.WriteJSON(w, http.StatusOK, s.customers.List())
}

func (s *Server) setCustomerRxPermit(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Permit bool `json:"permit"`
	}
	if err := platform.DecodeJSON(r, &body); err != nil {
		platform.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	updated, err := s.customers.SetRxPermit(pathID(r, "id"), body.Permit)
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, updated)
}
