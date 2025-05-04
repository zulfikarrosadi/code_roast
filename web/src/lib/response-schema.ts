export type Response<TData, TError> = {
  status: "success",
  code: number,
  data: TData
} | {
  status: "fail",
  code: number,
  error: {
    message: string,
    details?: TError,
  }
}
