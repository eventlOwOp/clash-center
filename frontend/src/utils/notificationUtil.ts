import Swal from 'sweetalert2'

export const notificationUtil = {
  // 成功消息提示
  showSuccessMessage(title: string, text: string) {
    return Swal.fire({
      title,
      text,
      icon: 'success',
      confirmButtonColor: '#3085d6',
      confirmButtonText: '确定'
    })
  },

  // 错误消息提示
  showErrorMessage(title: string, text: string) {
    return Swal.fire({
      title,
      text,
      icon: 'error',
      confirmButtonColor: '#d33',
      confirmButtonText: '确定'
    })
  },

  // 警告消息提示
  showWarningMessage(title: string, text: string) {
    return Swal.fire({
      title,
      text,
      icon: 'warning',
      confirmButtonColor: '#f8bb86',
      confirmButtonText: '确定'
    })
  },

  // 确认对话框
  confirmAction(title: string, text: string, confirmButtonText = '确定') {
    return Swal.fire({
      title,
      text,
      icon: 'question',
      showCancelButton: true,
      confirmButtonColor: '#3085d6',
      cancelButtonColor: '#d33',
      confirmButtonText,
      cancelButtonText: '取消'
    })
  }
} 