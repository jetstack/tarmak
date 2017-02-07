# class kubernetes::worker
class kubernetes::worker (
){
  include ::kubernetes::kubelet
}
