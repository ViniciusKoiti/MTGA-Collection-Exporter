"""
Aba "Configuração": caminhos, fonte do banco e gerenciamento das âncoras.

Implementada como um mixin consumido por gui.app.App — os atributos de estado
(self.config_obj, self.name_to_id, self._set_status, self.tabs...) vêm do App.
"""

from __future__ import annotations

import difflib
from tkinter import filedialog

import customtkinter as ctk


class ConfigTabMixin:
    # ------------------------------------------------------------ construção --
    def _build_config_tab(self):
        t = self.tab_config
        t.grid_columnconfigure(0, weight=1)

        # --- Caminhos -------------------------------------------------------
        paths = ctk.CTkFrame(t)
        paths.grid(row=0, column=0, padx=12, pady=12, sticky="ew")
        paths.grid_columnconfigure(1, weight=1)

        ctk.CTkLabel(paths, text="Caminhos", font=("", 16, "bold")).grid(
            row=0, column=0, columnspan=3, padx=12, pady=(12, 8), sticky="w"
        )
        ctk.CTkButton(
            paths,
            text="?",
            width=34,
            command=self._show_tutorial,
        ).grid(
            row=0,
            column=3,
            padx=(0, 12),
            pady=(12, 8),
            sticky="e",
        )

        ctk.CTkLabel(paths, text="Pasta 'Raw' do MTGA:").grid(
            row=1, column=0, padx=12, pady=6, sticky="w"
        )
        self.entry_mtga = ctk.CTkEntry(
            paths, placeholder_text="vazio = detectar automaticamente"
        )
        self.entry_mtga.grid(row=1, column=1, padx=6, pady=6, sticky="ew")
        ctk.CTkButton(
            paths, text="📁 Procurar", width=110, command=self._browse_mtga
        ).grid(row=1, column=2, padx=(6, 6), pady=6)
        ctk.CTkButton(
            paths, text="Detectar", width=90, command=self._detect_mtga
        ).grid(row=1, column=3, padx=(0, 12), pady=6)

        ctk.CTkLabel(paths, text="Pasta de saída:").grid(
            row=2, column=0, padx=12, pady=6, sticky="w"
        )
        self.entry_output = ctk.CTkEntry(
            paths, placeholder_text="vazio = pasta do programa"
        )
        self.entry_output.grid(row=2, column=1, padx=6, pady=6, sticky="ew")
        ctk.CTkButton(
            paths, text="📁 Procurar", width=110, command=self._browse_output
        ).grid(row=2, column=2, columnspan=2, padx=(6, 12), pady=6, sticky="w")

        ctk.CTkLabel(paths, text="Fonte do banco de cartas:").grid(
            row=3, column=0, padx=12, pady=(6, 12), sticky="w"
        )
        self.source_var = ctk.StringVar(value=self.config_obj.database_source)
        ctk.CTkSegmentedButton(
            paths, values=["auto", "local", "scryfall"], variable=self.source_var
        ).grid(row=3, column=1, padx=6, pady=(6, 12), sticky="w")

        # --- Âncoras --------------------------------------------------------
        anchors = ctk.CTkFrame(t)
        anchors.grid(row=1, column=0, padx=12, pady=(0, 12), sticky="nsew")
        anchors.grid_columnconfigure(0, weight=1)
        t.grid_rowconfigure(1, weight=1)

        ctk.CTkLabel(
            anchors, text="Âncoras de calibração", font=("", 16, "bold")
        ).grid(row=0, column=0, padx=12, pady=(12, 2), sticky="w")
        ctk.CTkLabel(
            anchors,
            text="Cartas raras/míticas que você possui, usadas para localizar a "
            "coleção na memória. Informe o nome e a quantidade exata.",
            font=("", 11),
            text_color="#9aa0a6",
            wraplength=760,
            justify="left",
        ).grid(row=1, column=0, padx=12, pady=(0, 8), sticky="w")

        add_row = ctk.CTkFrame(anchors, fg_color="transparent")
        add_row.grid(row=2, column=0, padx=8, pady=4, sticky="ew")
        add_row.grid_columnconfigure(0, weight=1)
        self.entry_anchor_name = ctk.CTkEntry(
            add_row, placeholder_text="Nome da carta (ex.: Sheoldred, the Apocalypse)"
        )
        self.entry_anchor_name.grid(row=0, column=0, padx=4, sticky="ew")
        self.entry_anchor_name.bind(
            "<Return>", lambda e: self.entry_anchor_qty.focus()
        )
        self.entry_anchor_qty = ctk.CTkEntry(add_row, width=70, placeholder_text="Qtd")
        self.entry_anchor_qty.grid(row=0, column=1, padx=4)
        self.entry_anchor_qty.bind("<Return>", lambda e: self._add_anchor())
        ctk.CTkButton(
            add_row, text="＋ Adicionar", width=120, command=self._add_anchor
        ).grid(row=0, column=2, padx=4)

        self.anchors_list = ctk.CTkScrollableFrame(anchors, height=180)
        self.anchors_list.grid(row=3, column=0, padx=8, pady=8, sticky="nsew")
        self.anchors_list.grid_columnconfigure(0, weight=1)
        anchors.grid_rowconfigure(3, weight=1)

        ctk.CTkButton(
            t, text="💾 Salvar configuração", command=self._save_config
        ).grid(row=2, column=0, padx=12, pady=(0, 12), sticky="e")

    # ------------------------------------------------------------- estado -----
    def _load_config_into_ui(self):
        self.entry_mtga.insert(0, self.config_obj.mtga_path)
        self.entry_output.insert(0, self.config_obj.output_folder)
        self.source_var.set(self.config_obj.database_source)
        self._render_anchors()

    def _render_anchors(self):
        for w in self.anchors_list.winfo_children():
            w.destroy()
        if not self.config_obj.anchors:
            ctk.CTkLabel(
                self.anchors_list,
                text="Nenhuma âncora ainda. Adicione ao menos 1 (idealmente 3-5).",
                text_color="#9aa0a6",
            ).grid(row=0, column=0, padx=8, pady=8, sticky="w")
            return
        for i, anchor in enumerate(self.config_obj.anchors):
            _, qty, name = anchor[0], anchor[1], anchor[2]
            row = ctk.CTkFrame(self.anchors_list, fg_color="#333333")
            row.grid(row=i, column=0, padx=4, pady=3, sticky="ew")
            row.grid_columnconfigure(0, weight=1)
            ctk.CTkLabel(row, text=f"{name}", anchor="w").grid(
                row=0, column=0, padx=10, pady=6, sticky="w"
            )
            ctk.CTkLabel(row, text=f"x{qty}", width=40).grid(row=0, column=1, padx=6)
            ctk.CTkButton(
                row,
                text="✕",
                width=32,
                fg_color="#7a3030",
                hover_color="#963c3c",
                command=lambda idx=i: self._remove_anchor(idx),
            ).grid(row=0, column=2, padx=(6, 8), pady=4)

    # ------------------------------------------------------------- âncoras ----
    def _add_anchor(self):
        name = self.entry_anchor_name.get().strip()
        qty_raw = self.entry_anchor_qty.get().strip()
        if not name:
            self._set_status("Informe o nome da carta.")
            return
        try:
            qty = int(qty_raw)
            if qty < 1:
                raise ValueError
        except ValueError:
            self._set_status("Quantidade inválida.")
            return

        if not self.name_to_id:
            self._set_status("Banco ainda carregando... tente em instantes.")
            return

        search = name.lower()
        cid = self.name_to_id.get(search)
        final_name = name
        if not cid:
            matches = difflib.get_close_matches(
                search, self.name_to_id.keys(), n=8, cutoff=0.5
            )
            if not matches:
                self._set_status(f"Carta '{name}' não encontrada.")
                return
            if len(matches) == 1:
                final_name = matches[0].title()
                cid = self.name_to_id[matches[0]]
            else:
                self._choose_match_dialog(matches, qty)
                return

        self.config_obj.anchors.append([cid, qty, final_name])
        self.entry_anchor_name.delete(0, "end")
        self.entry_anchor_qty.delete(0, "end")
        self._render_anchors()
        self._set_status(f"Âncora adicionada: {final_name} x{qty}")

    def _choose_match_dialog(self, matches, qty):
        dlg = ctk.CTkToplevel(self)
        dlg.title("Escolha a carta")
        dlg.geometry("420x360")
        dlg.transient(self)
        dlg.grab_set()
        ctk.CTkLabel(dlg, text="Você quis dizer?", font=("", 14, "bold")).pack(
            padx=16, pady=(16, 8)
        )
        scroll = ctk.CTkScrollableFrame(dlg, width=380, height=240)
        scroll.pack(padx=16, pady=8, fill="both", expand=True)

        def pick(m):
            cid = self.name_to_id[m]
            self.config_obj.anchors.append([cid, qty, m.title()])
            self.entry_anchor_name.delete(0, "end")
            self.entry_anchor_qty.delete(0, "end")
            self._render_anchors()
            self._set_status(f"Âncora adicionada: {m.title()} x{qty}")
            dlg.destroy()

        for m in matches:
            ctk.CTkButton(
                scroll, text=m.title(), command=lambda mm=m: pick(mm)
            ).pack(padx=6, pady=4, fill="x")

    def _remove_anchor(self, idx):
        try:
            del self.config_obj.anchors[idx]
            self._render_anchors()
        except IndexError:
            pass

    # ------------------------------------------------------------- ajuda ------
    def _show_tutorial(self):
        dlg = ctk.CTkToplevel(self)
        dlg.title("Passo a passo")
        dlg.geometry("620x520")
        dlg.transient(self)
        dlg.grab_set()

        ctk.CTkLabel(
            dlg,
            text="Como escanear sua coleção",
            font=("", 18, "bold"),
        ).pack(padx=18, pady=(18, 8), anchor="w")

        text = ctk.CTkTextbox(dlg, wrap="word")
        text.pack(padx=18, pady=(0, 14), fill="both", expand=True)
        text.insert(
            "1.0",
            (
                "1. Configure a pasta Raw do MTGA\n"
                "   Use Detectar ou selecione a pasta que termina em "
                "MTGA_Data/Downloads/Raw.\n\n"
                "2. Configure a pasta de saida\n"
                "   Os arquivos mtga_collection.json, .csv e .txt serao salvos "
                "nessa pasta.\n\n"
                "3. Escolha a fonte do banco\n"
                "   Use auto. O app tenta ler o banco local do MTGA e usa o cache "
                "arena_id_lookup.json nas proximas vezes.\n\n"
                "4. Aguarde o banco carregar\n"
                "   Espere aparecer: Banco carregado: ... cartas. Pronto para "
                "escanear.\n\n"
                "5. Adicione ancoras de calibracao\n"
                "   Informe cartas que voce possui e a quantidade exata. Prefira "
                "3 a 5 raras ou miticas. Exemplo: Sheoldred, the Apocalypse x1.\n\n"
                "6. Salve a configuracao\n"
                "   Clique em Salvar configuracao para gravar config.json.\n\n"
                "7. Abra o MTG Arena\n"
                "   Deixe o jogo aberto na tela Decks. Se a leitura falhar, rode "
                "o app como administrador.\n\n"
                "8. Escaneie a colecao\n"
                "   Clique em Escanear Colecao. Ao terminar, veja a aba Colecao "
                "e confira os arquivos exportados na pasta de saida."
            ),
        )
        text.configure(state="disabled")

        ctk.CTkButton(dlg, text="Fechar", command=dlg.destroy).pack(
            padx=18, pady=(0, 18), anchor="e"
        )

    # ---------------------------------------------------------- persistência --
    def _collect_config(self):
        self.config_obj.mtga_path = self.entry_mtga.get().strip()
        self.config_obj.output_folder = self.entry_output.get().strip()
        self.config_obj.database_source = self.source_var.get()

    def _save_config(self):
        self._collect_config()
        self.config_obj.save()
        self._set_status("Configuração salva em config.json.")

    def _browse_mtga(self):
        d = filedialog.askdirectory(title="Selecione a pasta 'Raw' do MTGA")
        if d:
            self.entry_mtga.delete(0, "end")
            self.entry_mtga.insert(0, d)

    def _detect_mtga(self):
        self._collect_config()
        p = self.config_obj.resolve_mtga_path()
        if p:
            self.entry_mtga.delete(0, "end")
            self.entry_mtga.insert(0, str(p))
            self._set_status(f"Instalação encontrada: {p}")
        else:
            self._set_status("Instalação do MTGA não encontrada automaticamente.")

    def _browse_output(self):
        d = filedialog.askdirectory(title="Selecione a pasta de saída")
        if d:
            self.entry_output.delete(0, "end")
            self.entry_output.insert(0, d)
