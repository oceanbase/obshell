/*
 * Copyright (c) 2024 OceanBase.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package task

type Template struct {
	tailNode      *Node
	nodes         []*Node
	Name          string
	isMaintenance bool
}

func (template *Template) AddNode(node *Node) {
	if template.tailNode != nil {
		template.tailNode.AddDownstream(node)
		node.AddUpstream(template.tailNode)
	}
	template.tailNode = node
	template.nodes = append(template.nodes, node)
}

func (template *Template) GetNodes() []*Node {
	return template.nodes
}

func (template *Template) IsEmpty() bool {
	return len(template.nodes) == 0
}

func (template *Template) IsMaintenance() bool {
	return template.isMaintenance
}

type TemplateBuilder struct {
	Template *Template
}

func NewTemplateBuilder(name string) *TemplateBuilder {
	return &TemplateBuilder{Template: &Template{Name: name,
		isMaintenance: true}}
}

func (builder *TemplateBuilder) Build() *Template {
	return builder.Template
}

func (builder *TemplateBuilder) AddNode(node *Node) *TemplateBuilder {
	builder.Template.AddNode(node)
	return builder
}

func (builder *TemplateBuilder) AddTask(task ExecutableTask, parallel bool) *TemplateBuilder {
	builder.AddNode(NewNode(task, parallel))
	return builder
}

func (builder *TemplateBuilder) AddTemplate(template *Template) *TemplateBuilder {
	for _, node := range template.nodes {
		node.downStream = nil
		node.upStream = nil
		builder.AddNode(node)
	}
	builder.Template.isMaintenance = builder.Template.isMaintenance || template.isMaintenance
	return builder
}

func (builder *TemplateBuilder) SetMaintenance(isMaintenance bool) *TemplateBuilder {
	builder.Template.isMaintenance = isMaintenance
	return builder
}
