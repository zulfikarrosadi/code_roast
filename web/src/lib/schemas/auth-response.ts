export type AuthErrorDetails = {
  email: string
  password: string
  fullname: string
}

export type AuthResponse = {
  user: {
    id: string
    fullname: string
    email: string
    roles: {
      id: number
      name: string
    }[]
  }
  access_token: string
  refresh_token: string
}
