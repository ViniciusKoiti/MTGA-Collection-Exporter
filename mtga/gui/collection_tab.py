"""
Aba "Coleção": tabela das cartas com busca e filtro por raridade.

Mixin consumido por gui.app.App.
"""

from __future__ import annotations

import subprocess
import sys
from tkinter import ttk

import customtkinter as ctk

from .. import collection as collection_mod
from .theme import RARITY_COLORS, style_treeview

MANA_ICONS = {
    "W": "⚪",
    "U": "🔵",
    "B": "⚫",
    "R": "🔴",
    "G": "🟢",
}


class CollectionTabMixin:
    def _build_collection_tab(self):
        t = self.tab_collection
        t.grid_columnconfigure(0, weight=1)
        t.grid_rowconfigure(1, weight=1)

        bar = ctk.CTkFrame(t, fg_color="transparent")
        bar.grid(row=0, column=0, padx=8, pady=8, sticky="ew")
        bar.grid_columnconfigure(0, weight=1)

        self.entry_search = ctk.CTkEntry(
            bar, placeholder_text="🔍  Buscar por nome, tipo ou set..."
        )
        self.entry_search.grid(row=0, column=0, padx=4, sticky="ew")
        self.entry_search.bind("<KeyRelease>", lambda e: self._refresh_table())

        self.filter_rarity = ctk.CTkOptionMenu(
            bar,
            values=["Todas", "mythic", "rare", "uncommon", "common", "basic"],
            command=lambda _: self._refresh_table(),
            width=130,
        )
        self.filter_rarity.set("Todas")
        self.filter_rarity.grid(row=0, column=1, padx=4)

        ctk.CTkButton(
            bar, text="📂 Abrir pasta", width=120, command=self._reveal_output
        ).grid(row=0, column=2, padx=4)

        table_frame = ctk.CTkFrame(t)
        table_frame.grid(row=1, column=0, padx=8, pady=(0, 8), sticky="nsew")
        table_frame.grid_columnconfigure(0, weight=1)
        table_frame.grid_rowconfigure(0, weight=1)

        style_treeview()
        cols = ("count", "name", "set", "cn", "rarity", "colors", "type")
        self.tree = ttk.Treeview(
            table_frame, columns=cols, show="headings", style="Mtga.Treeview"
        )
        headers = {
            "count": ("Qtd", 55),
            "name": ("Nome", 300),
            "set": ("Set", 70),
            "cn": ("Nº", 60),
            "rarity": ("Raridade", 100),
            "colors": ("Cores", 120),
            "type": ("Tipo", 220),
        }
        for col, (label, width) in headers.items():
            self.tree.heading(col, text=label)
            anchor = "center" if col in ("count", "set", "cn") else "w"
            self.tree.column(col, width=width, anchor=anchor)
        self.tree.grid(row=0, column=0, sticky="nsew")

        vsb = ttk.Scrollbar(table_frame, orient="vertical", command=self.tree.yview)
        vsb.grid(row=0, column=1, sticky="ns")
        self.tree.configure(yscrollcommand=vsb.set)

        self.lbl_summary = ctk.CTkLabel(
            t, text="Nenhuma coleção carregada ainda.", font=("", 12)
        )
        self.lbl_summary.grid(row=2, column=0, padx=12, pady=(0, 2), sticky="w")

        self.lbl_validation = ctk.CTkLabel(t, text="", font=("", 12), anchor="w")
        self.lbl_validation.grid(row=3, column=0, padx=12, pady=(0, 8), sticky="ew")

        # Tenta carregar uma exportação anterior.
        prev = collection_mod.load_exported_collection(
            self.config_obj.resolve_output_folder()
        )
        if prev:
            self.collection = prev
            self._refresh_table()

    def _refresh_table(self):
        query = self.entry_search.get().strip().lower()
        rarity_filter = self.filter_rarity.get()

        for item in self.tree.get_children():
            self.tree.delete(item)

        shown = 0
        total = 0
        for c in self.collection:
            if not self._matches_table_filters(c, query, rarity_filter):
                continue
            colors = self._format_colors(c.get("colors", []))
            rarity = c.get("rarity", "")
            self.tree.insert(
                "",
                "end",
                values=(
                    c["count"],
                    c["name"],
                    c.get("set", ""),
                    c.get("collector_number", ""),
                    rarity,
                    colors,
                    c.get("type_line", ""),
                ),
                tags=(rarity or "common",),
            )
            shown += 1
            total += c["count"]

        for r, col in RARITY_COLORS.items():
            self.tree.tag_configure(r or "common", foreground=col)

        self.lbl_summary.configure(
            text=f"Exibindo {shown} cartas ({total} cópias)"
            + (f" · filtro: {rarity_filter}" if rarity_filter != "Todas" else "")
        )
        self._refresh_validation_summary()

    def _refresh_validation_summary(self):
        report = getattr(self, "validation_report", None)
        if not report:
            self.lbl_validation.configure(text="")
            return

        suffix = ""
        if report.issues:
            suffix = " · " + " | ".join(issue.message for issue in report.issues[:2])
        self.lbl_validation.configure(text=report.summary + suffix)

    def _matches_table_filters(
        self, card: dict, query: str, rarity_filter: str
    ) -> bool:
        if rarity_filter != "Todas" and card.get("rarity", "") != rarity_filter:
            return False
        if not query:
            return True
        haystack = (
            f"{card['name']} {card.get('set', '')} "
            f"{card.get('type_line', '')}"
        ).lower()
        return query in haystack

    def _format_colors(self, colors: list) -> str:
        icons = [MANA_ICONS.get(color, color) for color in (colors or [])]
        return " ".join(icons) if icons else "◇"

    def _reveal_output(self):
        folder = self.config_obj.resolve_output_folder()
        try:
            if sys.platform.startswith("win"):
                subprocess.Popen(["explorer", str(folder)])
            elif sys.platform == "darwin":
                subprocess.Popen(["open", str(folder)])
            else:
                subprocess.Popen(["xdg-open", str(folder)])
        except Exception:
            self._set_status(f"Pasta: {folder}")
