package main

type PrepareFunc func(string) string

type Preparer interface {
	PrepareConfig(config *Config)
	PrepareCondition(condition *Condition)
	PrepareEndpoint(endpoint *Endpoint)
	PrepareSuccesses(success ...*Success)
	PrepareSuccess(success *Success)
	PrepareAuth(auth *Auth)
	PrepareHeaders(headers map[string][]string)
	Prepare(val string) string
}

type supplierPreparer struct {
	prepare PrepareFunc
}

func NewSuppliedPreparer(preparer PrepareFunc) Preparer {
	return &supplierPreparer{prepare: preparer}
}

func (o supplierPreparer) Prepare(val string) string {
	return o.Prepare(val)
}

func (o supplierPreparer) PrepareConfig(config *Config) {
	if config != nil {
		o.PrepareCondition(config.Cond)
		o.PrepareEndpoint(config.Endpoint)
		o.PrepareSuccesses(config.Success...)
	}
}

func (o supplierPreparer) PrepareCondition(condition *Condition) {
	if condition != nil {
		condition.Type = o.prepare(condition.Type)
		condition.Js = o.prepare(condition.Js)
	}
}

func (o supplierPreparer) PrepareEndpoint(endpoint *Endpoint) {
	if endpoint != nil {
		endpoint.Method = o.prepare(endpoint.Method)
		endpoint.Body = o.prepare(endpoint.Body)
		endpoint.Url = o.prepare(endpoint.Url)
		if endpoint.Auth != nil {
			o.PrepareAuth(endpoint.Auth)
		}
		if endpoint.Headers != nil {
			o.PrepareHeaders(endpoint.Headers)
		}
	}
}

func (o supplierPreparer) PrepareSuccesses(success ...*Success) {
	if success != nil && len(success) > 0 {
		for i := range success {
			o.PrepareSuccess(success[i])
		}
	}
}

func (o supplierPreparer) PrepareSuccess(success *Success) {
	if success != nil {
		success.Type = o.prepare(success.Type)
		success.Message = o.prepare(success.Message)
		o.PrepareEndpoint(success.Endpoint)
		o.PrepareConfig(success.Watcher)
		o.PrepareCondition(success.Cond)
	}
}

func (o supplierPreparer) PrepareAuth(auth *Auth) {
	if auth != nil {
		auth.Username = o.prepare(auth.Username)
		auth.Password = o.prepare(auth.Password)
		auth.Type = o.prepare(auth.Type)
	}
}

func (o supplierPreparer) PrepareHeaders(headers map[string][]string) {
	if headers != nil {
		for key, element := range headers {
			result := make([]string, len(element))
			for i := range element {
				result[i] = o.prepare(element[i])
			}
			headers[o.prepare(key)] = result
		}
	}
}
