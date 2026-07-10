"""
Janela principal do app: monta o layout, gerencia o banco em segundo plano e
conduz o scan numa thread. Reúne as abas via mixins (config_tab, collection_tab).
"""

from __future__ import annotations

import threading

import customtkinter as ctk

from .. import config as config_mod
from .. import database as database_mod
from ..collection import load_exported_collection
from ..i18n import translate
from ..pipeline import run_full_scan
from ..scanner import MtgaNotRunning, ScanError
from ..validation import export_validation_report, validate_collection
from .collection_tab import CollectionTabMixin
from .config_tab import ConfigTabMixin
from .theme import setup_appearance


class App(ctk.CTk, ConfigTabMixin, CollectionTabMixin):
    def __init__(self):
        super().__init__()
        self.title("MTGA Collection Exporter")
        self.geometry("1040x720")
        self.minsize(900, 600)

        self.config_obj = config_mod.Config.load()
        self.language = self.config_obj.language
        self.db: dict = {}
        self.name_to_id: dict = {}
        self.collection: list = []
        self.validation_report = None
        self._scanning = False

        self._build_layout()
        self._load_config_into_ui()

        threading.Thread(target=self._load_db_async, daemon=True).start()

    # ------------------------------------------------------------- layout -----
    def _build_layout(self):
        self.grid_columnconfigure(0, weight=1)
        self.grid_rowconfigure(0, weight=1)

        self.tabs = ctk.CTkTabview(self)
        self.tabs.grid(row=0, column=0, padx=16, pady=(16, 8), sticky="nsew")
        self._build_tabs()
        self._build_bottom_bar()

    def _build_tabs(self):
        self.tab_config = self.tabs.add(self._t("tab_config"))
        self.tab_collection = self.tabs.add(self._t("tab_collection"))

        self._build_config_tab()
        self._build_collection_tab()

    def _build_bottom_bar(self):
        bottom = ctk.CTkFrame(self)
        bottom.grid(row=1, column=0, padx=16, pady=(0, 16), sticky="ew")
        bottom.grid_columnconfigure(0, weight=1)

        self.progress = ctk.CTkProgressBar(bottom)
        self.progress.set(0)
        self.progress.grid(row=0, column=0, padx=12, pady=12, sticky="ew")

        self.lbl_status = ctk.CTkLabel(
            bottom, text=self._t("app_ready"), width=260, anchor="w"
        )
        self.lbl_status.grid(row=0, column=1, padx=8, pady=12)

        self.btn_scan = ctk.CTkButton(
            bottom,
            text=self._t("scan_collection"),
            width=200,
            height=40,
            font=("", 14, "bold"),
            command=self._start_scan,
        )
        self.btn_scan.grid(row=0, column=2, padx=12, pady=12)

    # -------------------------------------------------------------- banco -----
    def _load_db_async(self):
        self.db = database_mod.load_card_database(self.config_obj, self._progress_cb)
        self.name_to_id = database_mod.name_to_id_map(self.db)
        self.after(
            0,
            lambda: self._set_status(self._t("db_loaded", count=len(self.db))),
        )
        self.after(0, lambda: self.progress.set(0))

    # --------------------------------------------------------------- scan -----
    def _start_scan(self):
        if self._scanning:
            return
        self._collect_config()
        self.config_obj.save()
        if not self.config_obj.anchors:
            self.tabs.set(self._t("tab_config"))
            self._set_status(self._t("need_anchor"))
            return
        self._scanning = True
        self.btn_scan.configure(state="disabled", text=self._t("scanning"))
        threading.Thread(target=self._scan_worker, daemon=True).start()

    def _scan_worker(self):
        try:
            output_folder = self.config_obj.resolve_output_folder()
            previous = load_exported_collection(output_folder)
            final_list = run_full_scan(self.config_obj, self._progress_cb)
            report = validate_collection(
                final_list,
                previous_collection=previous,
                anchors=self.config_obj.anchors,
                database_size=len(self.db) if self.db else None,
            )
            export_validation_report(report, output_folder)
            self.collection = final_list
            self.validation_report = report
            self.after(0, self._on_scan_done, final_list, report)
        except (MtgaNotRunning, ScanError) as e:
            self.after(0, self._on_scan_error, str(e))
        except Exception as e:  # pragma: no cover
            self.after(0, self._on_scan_error, self._t("unexpected_error", error=e))

    def _on_scan_done(self, final_list, report):
        self._scanning = False
        self.btn_scan.configure(state="normal", text=self._t("scan_collection"))
        self.progress.set(1)
        total_cards = sum(i["count"] for i in final_list)
        self._set_status(
            self._t(
                "scan_done",
                status=self._status_label(report.status),
                unique=len(final_list),
                total=total_cards,
            )
        )
        self.tabs.set(self._t("tab_collection"))
        self._refresh_table()

    def _on_scan_error(self, msg):
        self._scanning = False
        self.btn_scan.configure(state="normal", text=self._t("scan_collection"))
        self.progress.set(0)
        self._set_status(f"⚠ {msg}")

    # ------------------------------------------------------------- helpers ----
    def _progress_cb(self, i, total, msg):
        frac = (i / total) if total else 0
        self.after(0, lambda: self.progress.set(min(1.0, max(0.0, frac))))
        self.after(0, lambda: self._set_status(msg))

    def _set_status(self, text):
        self.lbl_status.configure(text=text)

    def _apply_language(self, language):
        self._collect_config()
        self.language = language
        self.config_obj.language = language
        self.config_obj.save()
        self.tabs.destroy()
        self.tabs = ctk.CTkTabview(self)
        self.tabs.grid(row=0, column=0, padx=16, pady=(16, 8), sticky="nsew")
        self._build_tabs()
        self._load_config_into_ui()
        self._refresh_table()
        self.lbl_status.configure(text=self._t("config_saved"))
        self.btn_scan.configure(text=self._t("scan_collection"))

    def _status_label(self, status):
        return self._t(f"status_{status}")

    def _t(self, key, **kwargs):
        return translate(self.language, key, **kwargs)


def main():
    setup_appearance()
    App().mainloop()
