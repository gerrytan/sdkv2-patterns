---
tocTitle: Top level ForceNew in list
tocSection: Schema Design
---

# Top level ForceNew in list

Given a list of blocks with `ForceNew: true` specified at the top, but not in nested attributes:

<!-- ref-start ref="internal/reference-provider/provider/resource_paper_plane.go#list-top-level-force-new-schema" lang="go" -->

```go
"stickers": {
	Type:        schema.TypeList,
	Optional:    true,
	MinItems:    1,
	Description: "Stickers for the paper plane.",
	ForceNew:    true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The identifier of the sticker.",
			},
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The description of the sticker.",
			},
		},
	},
},
```

<!-- ref-end -->

Update to any member of the list will be done in-place. Full resource replacement will only occur if the list size changes[^1].

For example, given the following configuration:

<!-- ref-start file="internal/reference-provider/provider/testdata/resource_paper_plane_test_stickers.tf" lang="terraform" -->

```terraform
resource "hello_paper_plane" "test" {
  name = "test-paper-plane-stickers"

  stickers {
    id          = "sunrise"
    description = "A golden sunrise"
  }

  stickers {
    id          = "stars"
    description = "A sky full of stars"
  }
}
```

<!-- ref-end -->

Changing the id or description of existing sticker will be done in-place, but if new sticker are added or existing ones removed -- causing list size change, full resource replacement will occur.

- [^1]: Refer to [TestAccPaperPlaneStickersTopLevelOnlyForceNew](../internal/reference-provider/provider/resource_paper_plane_test.go)