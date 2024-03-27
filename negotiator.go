package vira

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// ResponseProcessor interface creates the contract for custom content negotiation.
type ResponseProcessor interface {
	CanProcess(mediaRange string) bool
	Process(w http.ResponseWriter, req *http.Request, dataModel interface{}, context ...interface{}) error
}

// AjaxResponseProcessor interface allows content negotiation to be biased when
// Ajax requests are handled. If a ResponseProcessor also implements this interface
// and its method returns true, then all Ajax requests will be fulfilled by that
// request processor, instead of via the normal content negotiation.
type AjaxResponseProcessor interface {
	IsAjaxResponder() bool
}

// WeightedValue is a value and associate weight between 0.0 and 1.0
type weightedValue struct {
	Value  string
	Weight float64
}

// ByWeight implements sort.Interface for []WeightedValue based
// on the Weight field. The data will be returned sorted decending
type byWeight []weightedValue

// Accept is an http accept
type accept string

type jsonProcessor struct {
	dense          bool
	prefix, indent string
	contentType    string
}

type xmlProcessor struct {
	dense          bool
	prefix, indent string
	contentType    string
}

const (
	xRequestedWith         = "X-Requested-With"
	xmlHttpRequest         = "XMLHttpRequest"
	defaultJSONContentType = "application/json"
	defaultXMLContentType  = "application/xml"
	// ParameteredMediaRangeWeight is the default weight of a media range with an
	// accept-param
	ParameteredMediaRangeWeight float64 = 1.0 //e.g text/html;level=1
	// TypeSubtypeMediaRangeWeight is the default weight of a media range with
	// type and subtype defined
	TypeSubtypeMediaRangeWeight float64 = 0.9 //e.g text/html
	// TypeStarMediaRangeWeight is the default weight of a media range with a type
	// defined but * for subtype
	TypeStarMediaRangeWeight float64 = 0.8 //e.g text/*
	// StarStarMediaRangeWeight is the default weight of a media range with any
	// type or any subtype defined
	StarStarMediaRangeWeight float64 = 0.7 //e.g */*
)

// Negotiator is responsible for content negotiation when using custom response processors.
type Negotiator struct{ processors []ResponseProcessor }

// NewWithJSONAndXML allows users to pass custom response processors. By default, processors
// for XML and JSON are already created.
func NewWithJSONAndXML(responseProcessors ...ResponseProcessor) *Negotiator {
	return NewNegotiator(append(responseProcessors, NewJSON(), NewXML())...)
}

// NewNegotiator allows users to pass custom response processors.
func NewNegotiator(responseProcessors ...ResponseProcessor) *Negotiator {
	return &Negotiator{
		responseProcessors,
	}
}

// Add more response processors. A new Negotiator is returned with the original processors plus
// the extra processors.
func (n *Negotiator) Add(responseProcessors ...ResponseProcessor) *Negotiator {
	return &Negotiator{
		append(n.processors, responseProcessors...),
	}
}

// Negotiate your model based on the HTTP Accept header.
func (n *Negotiator) Negotiate(w http.ResponseWriter, req *http.Request, dataModel interface{}, context ...interface{}) error {
	return negotiateHeader(n.processors, w, req, dataModel, context...)
}

// Negotiate your model based on the HTTP Accept header. Only XML and JSON are handled.
func Negotiate(w http.ResponseWriter, req *http.Request, dataModel interface{}, context ...interface{}) error {
	processors := []ResponseProcessor{NewJSON(), NewXML()}
	return negotiateHeader(processors, w, req, dataModel, context...)
}

// Firstly, all Ajax requests are processed by the first available Ajax processor.
// Otherwise, standard content negotiation kicks in.
//
// A request without any Accept header field implies that the user agent
// will accept any media type in response.
//
// If the header field is present in a request and none of the available
// representations for the response have a media type that is listed as
// acceptable, the origin server can either honour the header field by
// sending a 406 (Not Acceptable) response or disregard the header field
// by treating the response as if it is not subject to content negotiation.
// This implementation prefers the former.
//
// See rfc7231-sec5.3.2:
// http://tools.ietf.org/html/rfc7231#section-5.3.2
func negotiateHeader(processors []ResponseProcessor, w http.ResponseWriter, req *http.Request, dataModel interface{}, context ...interface{}) error {
	if IsAjax(req) {
		for _, processor := range processors {
			ajax, doesAjax := processor.(AjaxResponseProcessor)
			if doesAjax && ajax.IsAjaxResponder() {
				return processor.Process(w, req, dataModel, context...)
			}
		}
	}

	accept := accept(req.Header.Get("Accept"))

	if len(processors) > 0 {
		if accept == "" {
			return processors[0].Process(w, req, dataModel, context...)
		}

		for _, mr := range accept.ParseMediaRanges() {
			if len(mr.Value) == 0 {
				continue
			}

			if strings.EqualFold(mr.Value, "*/*") {
				return processors[0].Process(w, req, dataModel, context...)
			}

			for _, processor := range processors {
				if processor.CanProcess(mr.Value) {
					return processor.Process(w, req, dataModel, context...)
				}
			}
		}
	}

	http.Error(w, "", http.StatusNotAcceptable)
	return nil
}

// MediaRanges returns prioritized media ranges
func (accept accept) ParseMediaRanges() []weightedValue {
	var retVals []weightedValue
	mrs := strings.Split(string(accept), ",")

	for _, mr := range mrs {
		mrAndAcceptParam := strings.Split(mr, ";")
		//if no accept-param
		if len(mrAndAcceptParam) == 1 {
			retVals = append(retVals, handleMediaRangeNoAcceptParams(mrAndAcceptParam[0]))
			continue
		}

		retVals = append(retVals, handleMediaRangeWithAcceptParams(mrAndAcceptParam[0], mrAndAcceptParam[1:]))
	}

	//If no Accept header field is present, then it is assumed that the client
	//accepts all media types. If an Accept header field is present, and if the
	//server cannot send a response which is acceptable according to the combined
	//Accept field value, then the server SHOULD send a 406 (not acceptable)
	//response.
	sort.Sort(byWeight(retVals))

	return retVals
}

// IsAjax tests whether a request has the Ajax header.
func IsAjax(req *http.Request) bool {
	xRequestedWith, ok := req.Header[xRequestedWith]
	return ok && len(xRequestedWith) == 1 && xRequestedWith[0] == xmlHttpRequest
}

// NewJSON creates a new processor for JSON without indentation.
func NewJSON() ResponseProcessor {
	return &jsonProcessor{true, "", "", defaultJSONContentType}
}

func (*jsonProcessor) CanProcess(mediaRange string) bool {
	return strings.EqualFold(mediaRange, "application/json") ||
		strings.HasPrefix(mediaRange, "application/json-") ||
		strings.HasSuffix(mediaRange, "+json")
}

func (p *jsonProcessor) Process(w http.ResponseWriter, req *http.Request, dataModel interface{}, context ...interface{}) error {
	if dataModel == nil {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	w.Header().Set("Content-Type", p.contentType)
	if p.dense {
		return json.NewEncoder(w).Encode(dataModel)
	}

	js, err := json.MarshalIndent(dataModel, p.prefix, p.indent)

	if err != nil {
		return err
	}

	return writeWithNewline(w, js)
}

// NewXML creates a new processor for XML without indentation.
func NewXML() ResponseProcessor {
	return &xmlProcessor{true, "", "", defaultXMLContentType}
}

func (*xmlProcessor) CanProcess(mediaRange string) bool {
	return strings.Contains(mediaRange, "/xml") || strings.HasSuffix(mediaRange, "+xml")
}

func (p *xmlProcessor) Process(w http.ResponseWriter, req *http.Request, dataModel interface{}, context ...interface{}) error {
	if dataModel == nil {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	w.Header().Set("Content-Type", p.contentType)
	if p.dense {
		return xml.NewEncoder(w).Encode(dataModel)
	}

	x, err := xml.MarshalIndent(dataModel, p.prefix, p.indent)
	if err != nil {
		return err
	}

	return writeWithNewline(w, x)
}

func writeWithNewline(w io.Writer, x []byte) error {
	_, err := w.Write(x)
	if err != nil {
		return err
	}

	_, err = w.Write([]byte{'\n'})
	return err
}

func handleMediaRangeNoAcceptParams(mediaRange string) weightedValue {
	wv := new(weightedValue)
	wv.Value = strings.TrimSpace(mediaRange)
	wv.Weight = 0.0

	typeSubtype := strings.Split(wv.Value, "/")
	if len(typeSubtype) == 2 {
		switch {
		//a type of * with a non-star subtype is invalid, so if the type is
		//star the assume that the subtype is too
		case typeSubtype[0] == "*": //&& typeSubtype[1] == "*":
			wv.Weight = StarStarMediaRangeWeight
			break
		case typeSubtype[1] == "*":
			wv.Weight = TypeStarMediaRangeWeight
			break
		case typeSubtype[1] != "*":
			wv.Weight = TypeSubtypeMediaRangeWeight
			break
		}
	} //else invalid media range the weight remains 0.0

	return *wv
}

func handleMediaRangeWithAcceptParams(mediaRange string, acceptParams []string) weightedValue {
	wv := new(weightedValue)
	wv.Value = strings.TrimSpace(mediaRange)
	wv.Weight = ParameteredMediaRangeWeight

	for index := 0; index < len(acceptParams); index++ {
		ap := strings.ToLower(acceptParams[index])
		if isQualityAcceptParam(ap) {
			wv.Weight = parseQuality(ap)
		} else {
			wv.Value = strings.Join([]string{wv.Value, acceptParams[index]}, ";")
		}
	}
	return *wv
}

func isQualityAcceptParam(acceptParam string) bool {
	return strings.Contains(acceptParam, "q=")
}

func parseQuality(acceptParam string) float64 {
	weight, err := strconv.ParseFloat(strings.SplitAfter(acceptParam, "q=")[1], 64)
	if err != nil {
		weight = 1.0
	}
	return weight
}

func (a byWeight) Len() int           { return len(a) }
func (a byWeight) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byWeight) Less(i, j int) bool { return a[i].Weight > a[j].Weight }
