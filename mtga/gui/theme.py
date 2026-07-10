"""
Configuração de aparência e estilização da tabela (ttk.Treeview no tema escuro).
"""

from __future__ import annotations

import tkinter as tk
from tkinter import ttk

import customtkinter as ctk

RARITY_COLORS = {
    "mythic": "#f0a030",
    "rare": "#d4b45a",
    "uncommon": "#9fb0c0",
    "common": "#e0e0e0",
    "basic": "#e0e0e0",
    "": "#e0e0e0",
}


def setup_appearance() -> None:
    ctk.set_appearance_mode("dark")
    ctk.set_default_color_theme("blue")


def style_treeview() -> None:
    """Aplica o estilo 'Mtga.Treeview' combinando com o tema escuro do app."""
    style = ttk.Style()
    try:
        style.theme_use("default")
    except tk.TclError:
        pass
    style.configure(
        "Mtga.Treeview",
        background="#2b2b2b",
        foreground="#e0e0e0",
        fieldbackground="#2b2b2b",
        rowheight=26,
        borderwidth=0,
    )
    style.configure(
        "Mtga.Treeview.Heading",
        background="#1f1f1f",
        foreground="#c0c0c0",
        relief="flat",
        font=("", 11, "bold"),
    )
    style.map("Mtga.Treeview", background=[("selected", "#144870")])
