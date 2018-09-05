require 'spec_helper'
describe 'kubernetes_addons::metrics_server' do
  let(:pre_condition) do
    "
      class kubernetes{
        $_authorization_mode = ['RBAC']
        $version = '1.8.0'
      }
      define kubernetes::apply(
        $manifests,
      ){}
    "
  end

  let(:manifests) do
    catalogue.resource('Kubernetes::Apply', 'metrics-server').send(:parameters)[:manifests]
  end

  context 'with defaults' do
    it 'be valid yaml' do
      manifests.each do |manifest|
        YAML.parse manifest
      end
    end

    it 'have image set' do
      expect(manifests[0]).to match(%{^[-\s].*image: [^:]+:[^:]+$})
    end

    it 'have resources set' do
      expect(manifests[0]).to match(%{^[-\s].*cpu: [0-9]+})
      expect(manifests[0]).to match(%{^[-\s].*memory: [0-9]+})
    end

    it 'have nanny options set' do
      expect(manifests[0]).to match(%{^[-\s].*- --cpu=[0-9]+})
      expect(manifests[0]).to match(%{^[-\s].*- --memory=[0-9]+})
      expect(manifests[0]).to match(%{^[-\s].*- --extra-cpu=[0-9]+})
      expect(manifests[0]).to match(%{^[-\s].*- --extra-memory=[0-9]+})
    end
    context 'minikube tests' do
      after(:each) do
          kubectl_delete(manifests) if @minikube_cleanup
      end

      it 'deploy healthy app', :minikube => true do
        skip "minikube unstable"
        @minikube_cleanup = true
        expect(
          minikube_apply(manifests)
        ).to eq(0)

        retries = 10
        begin
          ready_replicas = kubectl_get('deployment','kube-system', 'metrics-server')['status']['readyReplicas']
          raise "deployment not healthy" if ready_replicas != 1
        rescue Exception => e
          retries -= 1
          raise e if retries == 0
          sleep 5
          retry
        end
      end
    end
  end
end
