package node

import (
	"flutelake/fluteNAS/pkg/model"
	"flutelake/fluteNAS/pkg/module/flog"
	"os"

	"github.com/flosch/pongo2"
)

func (x *Exec) StartNFSGanesha(nfs []model.NFSExport) error {
	// 安装应用
	bs, err := x.Command("apt install nfs-ganesha nfs-ganesha-vfs")
	if err != nil {
		flog.Errorf("install nfs-ganesha error, output: %s, err: %s", string(bs), err)
		return err
	}

	return x.generatehNFSGaneshaConfigAndStart(nfs)
}

func (x *Exec) StopNFSGanesha() error {
	bs, err := x.Command("systemctl disable nfs-ganesha && systemctl stop nfs-ganesha")
	if err != nil {
		flog.Errorf("start nfs-ganesha error, output: %s, err: %s", string(bs), err)
		return err
	}
	return nil
}

func (x *Exec) RefreshNFSGaneshaConfig(nfs []model.NFSExport) error {
	return x.generatehNFSGaneshaConfigAndStart(nfs)
}

func (x *Exec) generatehNFSGaneshaConfigAndStart(nfs []model.NFSExport) error {
	config := `NFS_CORE_PARAM {
        mount_path_pseudo = true;
        Protocols = 3,4,9P;
}

EXPORT_DEFAULTS {
        Access_Type = RW;
        Squash = all_squash;
		## todo replace uid gid
        Anonymous_Uid = 1000;
        Anonymous_Gid = 1000;
}

MDCACHE {
        Entries_HWMark = 100000;
}

LOG {

	Default_Log_Level = WARN;
	Components {
			FSAL = INFO;
			NFS4 = EVENT;
	}

	Facility {
			name = FILE;
			destination = "/var/log/ganesha.log";
			enable = active;
	}
}
{% for e in exports %}
EXPORT
{		
		# Export {{ e.Name }}
        Export_Id = {{ e.Id }};
        Path = {{ e.Path }};
        Pseudo = {{ e.Pseudo }};
        Protocols = 3,4;
        Access_Type = RW;
        FSAL {
                Name = VFS;
        }
        Clients = {{ e.IPWhiteRange }};
}
{% endfor %}
	`

	// Compile the template first (i. e. creating the AST)
	tpl, err := pongo2.FromString(config)
	if err != nil {
		flog.Errorf("Error compiling nfs export template: %v", err)
		return err
	}
	// Now you can render the template with the given
	// pongo2.Context how often you want to.
	out, err := tpl.Execute(pongo2.Context{"exports": nfs})
	if err != nil {
		panic(err)
	}
	// fmt.Println(out)

	// 写入配置
	err = os.WriteFile("/etc/ganesha/ganesha.conf", []byte(out), 0644)
	if err != nil {
		flog.Errorf("Error writing nfs export config: %v", err)
		return err
	}

	// 启动应用
	bs, err := x.Command("systemctl enable nfs-ganesha && systemctl restart nfs-ganesha")
	if err != nil {
		flog.Errorf("start nfs-ganesha error, output: %s, err: %s", string(bs), err)
		return err
	}

	return nil
}
