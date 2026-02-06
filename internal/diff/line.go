package diff

func parseLine(raw string, h *Hunk) Line {

	if len(raw) == 0 {
		return Line{Type: Context, Content: ""}
	}

	switch raw[0] {

	case '+':
		l := Line{
			Type:      Added,
			Content:   raw[1:],
			NewNumber: h.NewStart,
		}
		h.NewStart++
		return l

	case '-':
		l := Line{
			Type:      Removed,
			Content:   raw[1:],
			OldNumber: h.OldStart,
		}
		h.OldStart++
		return l

	default:
		l := Line{
			Type:      Context,
			Content:   raw[1:],
			OldNumber: h.OldStart,
			NewNumber: h.NewStart,
		}
		h.OldStart++
		h.NewStart++
		return l
	}
}
