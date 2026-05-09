/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package tools

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// BindUSBPorts handles the Freebox API quirk: when no USB port is bound, the
// API returns "" (empty string) instead of [] (empty array). A custom
// UnmarshalJSON keeps the Go-side type as a slice while tolerating both shapes.
type BindUSBPorts []string

func (b *BindUSBPorts) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) || bytes.Equal(trimmed, []byte(`""`)) {
		*b = nil
		return nil
	}
	var arr []string
	if err := json.Unmarshal(trimmed, &arr); err != nil {
		return err
	}
	*b = arr
	return nil
}

// VM reflects one entry from GET /api/v4/vm/
type VM struct {
	ID                int          `json:"id"`
	Name              string       `json:"name"`
	Status            string       `json:"status"`
	Memory            int          `json:"memory"`
	Vcpus             int          `json:"vcpus"`
	DiskPath          string       `json:"disk_path"`
	DiskType          string       `json:"disk_type"`
	OS                string       `json:"os"`
	EnableScreen      bool         `json:"enable_screen"`
	CloudinitEnabled  bool         `json:"cloudinit_enabled"`
	CloudinitUserdata string       `json:"cloudinit_userdata,omitempty"`
	CDPath            string       `json:"cd_path,omitempty"`
	BindUSBPorts      BindUSBPorts `json:"bind_usb_ports,omitempty"`
}

// vmCreateRequest is the body sent to POST /api/v4/vm/.
// It intentionally omits id and status (server-assigned fields): including
// them as zero-values causes the Freebox API to return invalid_request.
type vmCreateRequest struct {
	Name              string       `json:"name"`
	Memory            int          `json:"memory"`
	Vcpus             int          `json:"vcpus"`
	DiskPath          string       `json:"disk_path"`
	DiskType          string       `json:"disk_type"`
	OS                string       `json:"os"`
	EnableScreen      bool         `json:"enable_screen"`
	CloudinitEnabled  bool         `json:"cloudinit_enabled"`
	CloudinitUserdata string       `json:"cloudinit_userdata,omitempty"`
	CDPath            string       `json:"cd_path,omitempty"`
	BindUSBPorts      BindUSBPorts `json:"bind_usb_ports,omitempty"`
}

// maxCloudinitLen is the Freebox firmware limit for cloud-init userdata (Freebox bug FS#37547).
const maxCloudinitLen = 4096

// VMDiskTask reflects the async task returned by POST /api/v15/vm/disk/resize
// et POST /api/v15/vm/disk/create. Polled via GET /api/v15/vm/disk/task/{id}.
// Note : `error` est un bool (présence d'erreur), `error_message` est le texte
// associé. C'est différent de FSTask où `error` est une chaîne.
type VMDiskTask struct {
	ID           int    `json:"id"`
	Type         string `json:"type"`                    // create_disk | resize_disk
	State        string `json:"state"`                   // queued | running | done | failed
	Error        bool   `json:"error"`                   // true si erreur
	ErrorMessage string `json:"error_message,omitempty"` // message d'erreur si error=true
	Progress     int    `json:"progress"`                // 0..100
	DiskPath     string `json:"disk_path,omitempty"`
	Done         bool   `json:"done"`
	DurationLeft int    `json:"duration_left,omitempty"`
}

// vmDiskResizeRequest is the body sent to POST /api/v15/vm/disk/resize.
// disk_path is the existing disk's path, base64-encoded (same format as VM.DiskPath).
// shrink_allow doit être true uniquement si on diminue la taille (DESTRUCTIF).
type vmDiskResizeRequest struct {
	DiskPath    string `json:"disk_path"`
	Size        int64  `json:"size"`
	ShrinkAllow bool   `json:"shrink_allow"`
}

// vmDiskCreateRequest is the body sent to POST /api/v15/vm/disk/create.
// disk_path est le chemin absolu base64-encoded où créer le fichier.
type vmDiskCreateRequest struct {
	DiskPath string `json:"disk_path"`
	Size     int64  `json:"size"`
	DiskType string `json:"disk_type"` // qcow2 | raw
}

// bytesPerGB is the conversion factor used by the Freebox API (binary GB / GiB).
const bytesPerGB = 1024 * 1024 * 1024

func registerVM(s *server.MCPServer, c writer) {
	// ── Liste ────────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_vm_list",
			mcp.WithDescription("Liste les machines virtuelles de la Freebox : nom, état (running/stopped), mémoire, vCPUs, OS."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var vms []VM
			if err := c.Get(ctx, "/vm/", &vms); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(vms)
		},
	)

	// ── Démarrer ─────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_vm_start",
			mcp.WithDescription("Démarre une VM arrêtée (PRA : remontée de service)."),
			mcp.WithNumber("id",
				mcp.Required(),
				mcp.Description("Identifiant de la VM (voir freebox_vm_list)"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetInt("id", 0)
			if err := c.Post(ctx, fmt.Sprintf("/vm/%d/start", id), nil, nil); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf(`{"started": true, "id": %d}`, id)), nil
		},
	)

	// ── Arrêt gracieux (ACPI) ─────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_vm_stop",
			mcp.WithDescription("Envoie un signal d'arrêt ACPI à une VM (équivalent bouton Power). Préférer freebox_vm_kill si la VM ne répond plus."),
			mcp.WithNumber("id",
				mcp.Required(),
				mcp.Description("Identifiant de la VM (voir freebox_vm_list)"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetInt("id", 0)
			if err := c.Post(ctx, fmt.Sprintf("/vm/%d/stop", id), nil, nil); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf(`{"stopping": true, "id": %d}`, id)), nil
		},
	)

	// ── Forcer l'arrêt ───────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_vm_kill",
			mcp.WithDescription("Force l'arrêt d'une VM bloquée (équivalent coupure secteur). Risque de corruption des données."),
			mcp.WithNumber("id",
				mcp.Required(),
				mcp.Description("Identifiant de la VM (voir freebox_vm_list)"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetInt("id", 0)
			if err := c.Post(ctx, fmt.Sprintf("/vm/%d/kill", id), nil, nil); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf(`{"killed": true, "id": %d}`, id)), nil
		},
	)

	// ── Créer ────────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_vm_create",
			mcp.WithDescription("Crée une nouvelle machine virtuelle sur la Freebox."),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("Nom de la VM")),
			mcp.WithNumber("memory",
				mcp.Required(),
				mcp.Description("Mémoire allouée en Mo (ex: 2048)"),
				mcp.Min(64), mcp.Max(16384)),
			mcp.WithNumber("vcpus",
				mcp.Required(),
				mcp.Description("Nombre de vCPUs (1–8)"),
				mcp.Min(1), mcp.Max(8)),
			mcp.WithString("disk_name",
				mcp.Required(),
				mcp.Description("Nom du fichier disque (ex: fedora.qcow2, debian.raw) — extension .qcow2 ou .raw obligatoire"),
				mcp.Pattern(DiskNamePattern)),
			mcp.WithString("disk_dir",
				mcp.Required(),
				mcp.Description("Répertoire absolu du disque sur le stockage Freebox (ex: /Disque 1/VMs/, /Freebox/VMs/). Utiliser freebox_storage_partitions pour découvrir les chemins montés sur la Freebox cible. Encodé base64 en interne.")),
			mcp.WithString("disk_type",
				mcp.Required(),
				mcp.Description("Type de disque : raw ou qcow2"),
				mcp.Enum("qcow2", "raw")),
			mcp.WithString("os",
				mcp.Description("OS invité : fedora, debian, ubuntu, unknown (défaut: unknown)"),
				mcp.Enum("fedora", "debian", "ubuntu", "unknown")),
			mcp.WithBoolean("enable_screen",
				mcp.Description("Activer l'écran virtuel VNC (défaut: false)")),
			mcp.WithString("cloudinit_userdata",
				mcp.Description("YAML cloud-init injecté au boot (SSH keys, packages, runcmd…). Limite 4096 caractères (utiliser #include https://... pour les configs plus longues).")),
			mcp.WithString("cd_path",
				mcp.Description("Chemin absolu vers une ISO à monter (ex: /Disque 1/ISO/debian-12-arm64.iso). Encodé base64 en interne.")),
			mcp.WithArray("bind_usb_ports",
				mcp.Description("Ports USB à passer à la VM (ex: usb-external-type-c, usb-external-type-a)."),
				mcp.WithStringItems(),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name := req.GetString("name", "")
			memory := req.GetInt("memory", 0)
			vcpus := req.GetInt("vcpus", 0)
			diskName := req.GetString("disk_name", "")
			if err := validateDiskName(diskName); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			// Security by design: the caller provides the filename (disk_name) and
			// the directory (disk_dir). The directory is validated with sanitizeFSPath
			// to prevent path traversal attacks. disk_dir is required — no default —
			// because storage paths vary across Freebox models (e.g. /Disque 1/ vs
			// /Freebox/) and silent defaults mask configuration mismatches.
			diskDir := req.GetString("disk_dir", "")
			if diskDir == "" {
				return mcp.NewToolResultError("disk_dir : paramètre requis (utiliser freebox_storage_partitions pour lister les chemins disponibles)"), nil
			}
			cleanDir, err := sanitizeFSPath(diskDir)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("disk_dir : %v", err)), nil
			}
			diskPath := cleanDir + "/" + diskName
			diskType := req.GetString("disk_type", "")
			osName := req.GetString("os", "unknown")
			enableScreen := req.GetBool("enable_screen", false)

			body := vmCreateRequest{
				Name: name, Memory: memory, Vcpus: vcpus,
				DiskPath: base64.StdEncoding.EncodeToString([]byte(diskPath)),
				DiskType: diskType,
				OS:       osName, EnableScreen: enableScreen,
			}

			if userdata := req.GetString("cloudinit_userdata", ""); userdata != "" {
				if len(userdata) > maxCloudinitLen {
					return mcp.NewToolResultError(
						fmt.Sprintf("cloudinit_userdata dépasse la limite de %d caractères (utiliser #include https://... pour les configs plus longues)", maxCloudinitLen),
					), nil
				}
				body.CloudinitEnabled = true
				body.CloudinitUserdata = userdata
			}

			if cdRaw := req.GetString("cd_path", ""); cdRaw != "" {
				clean, err := sanitizeFSPath(cdRaw)
				if err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("cd_path : %v", err)), nil
				}
				body.CDPath = base64.StdEncoding.EncodeToString([]byte(clean))
			}

			if ports := req.GetStringSlice("bind_usb_ports", nil); len(ports) > 0 {
				body.BindUSBPorts = ports
			}

			var created VM
			if err := c.Post(ctx, "/vm/", body, &created); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(created)
		},
	)

	// ── Modifier ─────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_vm_update",
			mcp.WithDescription("Modifie la configuration d'une VM arrêtée (nom, mémoire, vCPUs, écran VNC)."),
			mcp.WithNumber("id",
				mcp.Required(),
				mcp.Description("Identifiant de la VM (voir freebox_vm_list)")),
			mcp.WithString("name",
				mcp.Description("Nouveau nom (optionnel)")),
			mcp.WithNumber("memory",
				mcp.Description("Nouvelle mémoire en Mo (optionnel)")),
			mcp.WithNumber("vcpus",
				mcp.Description("Nouveau nombre de vCPUs (optionnel)")),
			mcp.WithBoolean("enable_screen",
				mcp.Description("Activer/désactiver l'écran VNC (optionnel)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetInt("id", 0)
			// Read-modify-write : l'API Freebox PUT /vm/{id} rejette les patchs
			// partiels (#80). On lit l'état courant, on applique les overrides,
			// on renvoie le body complet — pattern utilisé par fbxvm-ctrl.
			var body VM
			if err := c.Get(ctx, fmt.Sprintf("/vm/%d", id), &body); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			args := req.GetArguments()
			if v, ok := args["name"].(string); ok && v != "" {
				body.Name = v
			}
			if v, ok := args["memory"]; ok && v != nil {
				body.Memory = int(toFloat(v))
			}
			if v, ok := args["vcpus"]; ok && v != nil {
				body.Vcpus = int(toFloat(v))
			}
			if v, ok := args["enable_screen"].(bool); ok {
				body.EnableScreen = v
			}
			var updated VM
			if err := c.Put(ctx, fmt.Sprintf("/vm/%d", id), body, &updated); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(updated)
		},
	)

	// ── Supprimer ────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_vm_delete",
			mcp.WithDescription("Supprime définitivement une VM et libère son disque virtuel."),
			mcp.WithNumber("id",
				mcp.Required(),
				mcp.Description("Identifiant de la VM à supprimer (voir freebox_vm_list)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetInt("id", 0)
			if err := c.Delete(ctx, fmt.Sprintf("/vm/%d", id)); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("VM %d supprimée.", id)), nil
		},
	)

	// ── Redimensionner le disque ─────────────────────────────────────────────
	// Cas typique : image cloud (~500 Mo) à étendre à la taille cible avant
	// premier boot (cloudinit cloudgrowfs grossira la partition au boot).
	s.AddTool(
		mcp.NewTool("freebox_vm_disk_resize",
			mcp.WithDescription("Redimensionne le disque qcow2 d'une VM. La VM doit être arrêtée. Réduire la taille (allow_shrink=true) est destructif. Opération asynchrone : retourne une tâche, à suivre avec freebox_vm_disk_task."),
			mcp.WithNumber("id",
				mcp.Required(),
				mcp.Description("Identifiant de la VM (voir freebox_vm_list)")),
			mcp.WithNumber("size_gb",
				mcp.Required(),
				mcp.Description("Nouvelle taille du disque en GiB (1 GiB = 1024^3 octets)")),
			mcp.WithBoolean("allow_shrink",
				mcp.Description("Autorise la réduction de taille (DESTRUCTIF — perte de données possible). Défaut: false")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetInt("id", 0)
			sizeGB := toFloat(req.GetArguments()["size_gb"])
			if sizeGB <= 0 {
				return mcp.NewToolResultError("size_gb doit être > 0"), nil
			}
			allowShrink := req.GetBool("allow_shrink", false)

			// Récupère le disk_path actuel (déjà base64) depuis la VM.
			var vm VM
			if err := c.Get(ctx, fmt.Sprintf("/vm/%d", id), &vm); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if vm.Status == "running" {
				return mcp.NewToolResultError(fmt.Sprintf("VM %d est en cours d'exécution — l'arrêter avant resize (freebox_vm_stop)", id)), nil
			}

			body := vmDiskResizeRequest{
				DiskPath:    vm.DiskPath,
				Size:        int64(sizeGB * float64(bytesPerGB)),
				ShrinkAllow: allowShrink,
			}
			var task VMDiskTask
			if err := c.Post(ctx, "/vm/disk/resize", body, &task); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(task)
		},
	)

	// ── Créer un disque qcow2/raw vide ───────────────────────────────────────
	// Utile en amont de freebox_vm_create quand on veut un disque "blank" plutôt
	// qu'une image cloud uploadée. Opération asynchrone (task à suivre via
	// freebox_vm_disk_task).
	s.AddTool(
		mcp.NewTool("freebox_vm_disk_create",
			mcp.WithDescription("Crée un fichier disque virtuel vide (qcow2 ou raw) sur le stockage Freebox. Pré-requis avant freebox_vm_create si l'image n'est pas déjà uploadée. Opération asynchrone : retourne une tâche, à suivre avec freebox_vm_disk_task."),
			mcp.WithString("disk_name",
				mcp.Required(),
				mcp.Description("Nom du fichier disque (ex: my-vm.qcow2)"),
				mcp.Pattern(DiskNamePattern)),
			mcp.WithString("disk_dir",
				mcp.Required(),
				mcp.Description("Répertoire absolu du disque sur le stockage Freebox (ex: /Disque 1/VMs/). Utiliser freebox_storage_partitions pour découvrir les chemins.")),
			mcp.WithNumber("size_gb",
				mcp.Required(),
				mcp.Description("Taille du disque en GiB (1 GiB = 1024^3 octets)")),
			mcp.WithString("disk_type",
				mcp.Required(),
				mcp.Description("Type : qcow2 ou raw"),
				mcp.Enum("qcow2", "raw")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			diskName := req.GetString("disk_name", "")
			if err := validateDiskName(diskName); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			diskDir := req.GetString("disk_dir", "")
			cleanDir, err := sanitizeFSPath(diskDir)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("disk_dir : %v", err)), nil
			}
			sizeGB := toFloat(req.GetArguments()["size_gb"])
			if sizeGB <= 0 {
				return mcp.NewToolResultError("size_gb doit être > 0"), nil
			}
			diskType := req.GetString("disk_type", "qcow2")

			body := vmDiskCreateRequest{
				DiskPath: base64.StdEncoding.EncodeToString([]byte(cleanDir + "/" + diskName)),
				Size:     int64(sizeGB * float64(bytesPerGB)),
				DiskType: diskType,
			}
			var task VMDiskTask
			if err := c.Post(ctx, "/vm/disk/create", body, &task); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(task)
		},
	)

	// ── Statut d'une tâche disque (resize, create_disk) ──────────────────────
	s.AddTool(
		mcp.NewTool("freebox_vm_disk_task",
			mcp.WithDescription("Récupère le statut d'une tâche disque VM (resize, create_disk). Utilisé pour suivre une opération longue après freebox_vm_disk_resize."),
			mcp.WithNumber("task_id",
				mcp.Required(),
				mcp.Description("Identifiant de la tâche (champ id retourné par freebox_vm_disk_resize)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			tid := req.GetInt("task_id", 0)
			var task VMDiskTask
			if err := c.Get(ctx, fmt.Sprintf("/vm/disk/task/%d", tid), &task); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(task)
		},
	)
}
