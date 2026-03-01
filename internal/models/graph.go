package models

type GraphNodeType string

type GraphEdgeType string

const (
	NodeTypePaper       GraphNodeType = "paper"
	NodeTypeAuthor      GraphNodeType = "author"
	NodeTypeConcept     GraphNodeType = "concept"
	NodeTypeInstitution GraphNodeType = "institution"
)

const (
	EdgeTypeAuthored   GraphEdgeType = "authored"
	EdgeTypeHasConcept GraphEdgeType = "has_concept"
	EdgeTypeCites      GraphEdgeType = "cites"
	EdgeTypeRelated    GraphEdgeType = "related"
	EdgeTypeAffiliated GraphEdgeType = "affiliated"
)

type GraphNode struct {
	ID         string         `json:"id"`
	Label      string         `json:"label"`
	Type       GraphNodeType  `json:"type"`
	Properties map[string]any `json:"properties,omitempty"`
}

type GraphEdge struct {
	Source       string        `json:"source"`
	Target       string        `json:"target"`
	Relationship GraphEdgeType `json:"relationship"`
	Weight       float64       `json:"weight"`
}

type KnowledgeGraph struct {
	Nodes        []GraphNode `json:"nodes"`
	Edges        []GraphEdge `json:"edges"`
	CenterNodeID string      `json:"center_node_id"`
}
