package layout

import "github.com/lxn/walk"

// GreedyLayoutItem is like walk.NewGreedyLayoutItem, but with specific orientation support.
// If an orientation is provided, it will be greedy at the given orientation but not other orientations.
type GreedyLayoutItem struct {
	*walk.LayoutItemBase
	item        walk.LayoutItem
	orientation walk.LayoutFlags
}

// NewGreedyLayoutItem returns a layout item that is greedy at the given orientation.
func NewGreedyLayoutItem(orientation walk.Orientation) walk.LayoutItem {
	layout := &GreedyLayoutItem{item: walk.NewGreedyLayoutItem()}
	layout.LayoutItemBase = layout.item.AsLayoutItemBase()
	switch orientation {
	case walk.Horizontal:
		layout.orientation = walk.GreedyVert
	case walk.Vertical:
		layout.orientation = walk.GreedyHorz
	case walk.NoOrientation:
		layout.orientation = walk.LayoutFlags(orientation)
	default:
		panic("invalid orientation")
	}
	return layout
}

func (hg *GreedyLayoutItem) LayoutFlags() walk.LayoutFlags {
	return hg.item.LayoutFlags() & ^hg.orientation
}

func (hg *GreedyLayoutItem) IdealSize() walk.Size {
	return hg.item.(walk.IdealSizer).IdealSize()
}

func (hg *GreedyLayoutItem) MinSize() walk.Size {
	return hg.item.(walk.MinSizer).MinSize()
}
