// for information about these interfaces
declare global {
  namespace App {
    interface Locals {
      user: {
        id: string;
        fullname: string;
        email: string;
      };
    }
  }
}

export {};
