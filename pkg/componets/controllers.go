package components

import (
	"github.com/openshift/microshift/pkg/assets"

	"github.com/sirupsen/logrus"
)

func startServiceCAController() error {
	var (
		//TODO: fix the rolebinding and sa
		clusterRoleBinding = []string{
			"assets/rbac/0000_60_service-ca_00_roles.yaml",
		}
		apps = []string{
			"assets/apps/0000_60_service-ca_05_deploy.yaml",
		}
		ns = []string{
			"assets/core/0000_60_service-ca_01_namespace.yaml",
		}
		sa = []string{
			"assets/core/0000_60_service-ca_04_sa.yaml",
		}
	)
	if err := assets.ApplyNamespaces(ns); err != nil {
		logrus.Warningf("failed to apply ns %v: %v", ns, err)
		return err
	}
	if err := assets.ApplyClusterRoleBindings(clusterRoleBinding); err != nil {
		logrus.Warningf("failed to apply clusterRolebinding %v: %v", clusterRoleBinding, err)
		return err
	}
	if err := assets.ApplyServiceAccounts(sa); err != nil {
		logrus.Warningf("failed to apply sa %v: %v", sa, err)
		return err
	}
	if err := assets.ApplyDeployments(apps, RenderSCController); err != nil {
		logrus.Warningf("failed to apply apps %v: %v", apps, err)
		return err
	}
	return nil
}

func startIngressController() error {
	var (
		clusterRoleBinding = []string{
			"assets/rbac/0000_80_openshift-router-cluster-role-binding.yaml",
		}
		clusterRole = []string{
			"assets/rbac/0000_80_openshift-router-cluster-role.yaml",
		}
		apps = []string{
			"assets/apps/0000_80_openshift-router-deployment.yaml",
		}
		ns = []string{
			"assets/core/0000_80_openshift-router-namespace.yaml",
		}
		sa = []string{
			"assets/core/0000_80_openshift-router-service-account.yaml",
		}
		cm = []string{
			"assets/core/0000_80_openshift-router-cm.yaml",
		}
		svc = []string{
			"assets/core/0000_80_openshift-router-service.yaml",
		}
	)
	if err := assets.ApplyNamespaces(ns); err != nil {
		logrus.Warningf("failed to apply ns %v: %v", ns, err)
		return err
	}
	if err := assets.ApplyClusterRoles(clusterRole); err != nil {
		logrus.Warningf("failed to apply clusterRolebinding %v: %v", clusterRole, err)
		return err
	}
	if err := assets.ApplyClusterRoleBindings(clusterRoleBinding); err != nil {
		logrus.Warningf("failed to apply clusterRolebinding %v: %v", clusterRoleBinding, err)
		return err
	}
	if err := assets.ApplyServiceAccounts(sa); err != nil {
		logrus.Warningf("failed to apply sa %v: %v", sa, err)
		return err
	}
	if err := assets.ApplyConfigMaps(cm); err != nil {
		logrus.Warningf("failed to apply cm %v: %v", cm, err)
		return err
	}
	if err := assets.ApplyServices(svc, nil); err != nil {
		logrus.Warningf("failed to apply svc %v: %v", svc, err)
		return err
	}
	if err := assets.ApplyDeployments(apps, nil); err != nil {
		logrus.Warningf("failed to apply apps %v: %v", apps, err)
		return err
	}
	return nil
}

func startDNSController() error {
	var (
		clusterRoleBinding = []string{
			"assets/rbac/0000_70_dns_01-cluster-role-binding.yaml",
		}
		clusterRole = []string{
			"assets/rbac/0000_70_dns_01-cluster-role.yaml",
		}
		apps = []string{
			"assets/apps/0000_70_dns_01-daemonset.yaml",
		}
		ns = []string{
			"assets/core/0000_70_dns_00-namespace.yaml",
		}
		sa = []string{
			"assets/core/0000_70_dns_01-service-account.yaml",
		}
		cm = []string{
			"assets/core/0000_70_dns_01-configmap.yaml",
		}
		svc = []string{
			"assets/core/0000_70_dns_01-service.yaml",
		}
	)
	if err := assets.ApplyNamespaces(ns); err != nil {
		logrus.Warningf("failed to apply ns %v: %v", ns, err)
		return err
	}
	if err := assets.ApplyServices(svc, render.RenderDNSService); err != nil {
		logrus.Warningf("failed to apply svc %v: %v", svc, err)
		// service already created by coreDNS, not re-create it.
		return nil
	}
	if err := assets.ApplyClusterRoles(clusterRole); err != nil {
		logrus.Warningf("failed to apply clusterRolebinding %v: %v", clusterRole, err)
		return err
	}
	if err := assets.ApplyClusterRoleBindings(clusterRoleBinding); err != nil {
		logrus.Warningf("failed to apply clusterRolebinding %v: %v", clusterRoleBinding, err)
		return err
	}
	if err := assets.ApplyServiceAccounts(sa); err != nil {
		logrus.Warningf("failed to apply sa %v: %v", sa, err)
		return err
	}
	if err := assets.ApplyConfigMaps(cm); err != nil {
		logrus.Warningf("failed to apply cm %v: %v", cm, err)
		return err
	}
	if err := assets.ApplyDaemonSets(apps, nil); err != nil {
		logrus.Warningf("failed to apply apps %v: %v", apps, err)
		return err
	}
	return nil
}