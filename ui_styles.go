package main

import (
	"github.com/charmbracelet/lipgloss"
)

type ThemeColors struct {
	Base    string
	Mantle  string
	Crust   string
	Text    string
	Subtext string
	Muted   string
	Overlay string
	Primary string
	Accent  string
	Success string
	Warning string
	Error   string
	Info    string
}

type UIStyles struct {
	Colors ThemeColors

	HeaderStyle       lipgloss.Style
	SidebarStyle      lipgloss.Style
	ChatViewportStyle lipgloss.Style
	InputStyle        lipgloss.Style
	InputActiveStyle  lipgloss.Style
	FooterStyle       lipgloss.Style
	
	ActiveTabStyle    lipgloss.Style
	InactiveTabStyle  lipgloss.Style
	
	ContactActiveStyle           lipgloss.Style
	ContactSelectedInactiveStyle lipgloss.Style
	ContactInactiveStyle         lipgloss.Style
	
	StatusOnlineStyle  lipgloss.Style
	StatusOfflineStyle lipgloss.Style
	
	ModalStyle       lipgloss.Style
	ModalHeaderStyle lipgloss.Style
}

func GetThemeColors(themeName string) ThemeColors {
	switch themeName {
	case "Nord":
		return ThemeColors{
			Base:    "#2e3440",
			Mantle:  "#242933",
			Crust:   "#1b1f27",
			Text:    "#d8dee9",
			Subtext: "#e5e9f0",
			Muted:   "#4c566a",
			Overlay: "#434c5e",
			Primary: "#88c0d0",
			Accent:  "#b48ead",
			Success: "#a3be8c",
			Warning: "#ebcb8b",
			Error:   "#bf616a",
			Info:    "#8fbcbb",
		}
	case "Gruvbox":
		return ThemeColors{
			Base:    "#282828",
			Mantle:  "#1d2021",
			Crust:   "#151718",
			Text:    "#ebdbb2",
			Subtext: "#a89984",
			Muted:   "#504945",
			Overlay: "#665c54",
			Primary: "#fabd2f",
			Accent:  "#fe8019",
			Success: "#b8bb26",
			Warning: "#d79921",
			Error:   "#fb4934",
			Info:    "#83a598",
		}
	case "Dracula":
		return ThemeColors{
			Base:    "#282a36",
			Mantle:  "#1e1f29",
			Crust:   "#191a21",
			Text:    "#f8f8f2",
			Subtext: "#6272a4",
			Muted:   "#44475a",
			Overlay: "#44475a",
			Primary: "#bd93f9",
			Accent:  "#ff79c6",
			Success: "#50fa7b",
			Warning: "#f1fa8c",
			Error:   "#ff5555",
			Info:    "#8be9fd",
		}
	case "Tokyo Night":
		return ThemeColors{
			Base:    "#1a1b26",
			Mantle:  "#16161e",
			Crust:   "#1f2335",
			Text:    "#a9b1d6",
			Subtext: "#787c99",
			Muted:   "#3b4261",
			Overlay: "#2e3c64",
			Primary: "#7aa2f7",
			Accent:  "#bb9af7",
			Success: "#9ece6a",
			Warning: "#e0af68",
			Error:   "#f7768e",
			Info:    "#7dcfff",
		}
	default: // "Catppuccin Mocha"
		return ThemeColors{
			Base:    "#1e1e2e",
			Mantle:  "#181825",
			Crust:   "#11111b",
			Text:    "#cdd6f4",
			Subtext: "#a6adc8",
			Muted:   "#313244",
			Overlay: "#45475a",
			Primary: "#b4befe",
			Accent:  "#cba6f7",
			Success: "#a6e3a1",
			Warning: "#f9e2af",
			Error:   "#f38ba8",
			Info:    "#89b4fa",
		}
	}
}

func GetStyles(themeName string) UIStyles {
	colors := GetThemeColors(themeName)

	return UIStyles{
		Colors: colors,

		HeaderStyle: lipgloss.NewStyle().
			Background(lipgloss.Color(colors.Crust)).
			Foreground(lipgloss.Color(colors.Text)).
			Bold(true).
			Height(1),

		SidebarStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colors.Muted)).
			Background(lipgloss.Color(colors.Mantle)).
			Padding(0, 1),

		ChatViewportStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colors.Muted)).
			Background(lipgloss.Color(colors.Base)).
			Padding(0, 1),

		InputStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colors.Muted)).
			Padding(0, 1),

		InputActiveStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colors.Primary)).
			Padding(0, 1),

		FooterStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Subtext)).
			Background(lipgloss.Color(colors.Crust)).
			Height(1),

		ActiveTabStyle: lipgloss.NewStyle().
			Background(lipgloss.Color(colors.Primary)).
			Foreground(lipgloss.Color(colors.Crust)).
			Bold(true).
			Padding(0, 1),

		InactiveTabStyle: lipgloss.NewStyle().
			Background(lipgloss.Color(colors.Muted)).
			Foreground(lipgloss.Color(colors.Text)).
			Padding(0, 1),

		ContactActiveStyle: lipgloss.NewStyle().
			Background(lipgloss.Color(colors.Primary)).
			Foreground(lipgloss.Color(colors.Crust)).
			Bold(true),

		ContactSelectedInactiveStyle: lipgloss.NewStyle().
			Background(lipgloss.Color(colors.Overlay)).
			Foreground(lipgloss.Color(colors.Text)).
			Bold(true),

		ContactInactiveStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Text)),

		StatusOnlineStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Success)).
			Bold(true),

		StatusOfflineStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Error)),

		ModalStyle: lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color(colors.Accent)).
			Background(lipgloss.Color(colors.Mantle)).
			Padding(1, 2).
			Align(lipgloss.Center),

		ModalHeaderStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Accent)).
			Bold(true).
			MarginBottom(1),
	}
}
