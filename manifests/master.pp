# class kubernetes::master
class kubernetes::master (
){
  include ::kubernetes::apiserver
  include ::kubernetes::controller_manager
  include ::kubernetes::scheduler
}
