require 'spec_helper'
describe 'kubernetes_addons::dashboard' do
  let(:pre_condition) do
    "
      class kubernetes{}
      define kubernetes::apply(
        $manifests,
      ){}
    "
  end

  let(:manifests) do
    catalogue.resource('Kubernetes::Apply', 'kube-dashboard').send(:parameters)[:manifests]
  end

  context 'with defaults' do
    it 'be valid yaml' do
      manifests.each do |manifest|
        YAML.parse manifest
      end
    end

    it 'have image set' do
      expect(manifests[1]).to match(%{^[-\s].*image: [^:]+:[^:]+$})
    end

    it 'have resources set' do
      expect(manifests[1]).to match(%{^[-\s].*cpu: [0-9]+})
      expect(manifests[1]).to match(%{^[-\s].*memory: [0-9]+})
    end

    context 'minikube tests' do
      after(:each) do
          kubectl_delete(manifests) if ENV['MINIKUBE'] == 'true'
      end

      it 'deploy healthy app', :minikube => true do
        skip "Minikube tests disabled by default, pass MINIKUBE=true as env var" if ENV['MINIKUBE'] != 'true'
        expect(
          minikube_apply(manifests)
        ).to eq(0)

        retries = 10
        begin
          ready_replicas = kubectl_get('deployment','kube-system', 'kubernetes-dashboard')['status']['readyReplicas']
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
