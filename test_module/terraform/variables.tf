variable "local_file_path" {
  type = string
}

variable "bool" {
  type = bool
}

variable "bool_with_default" {
  type    = bool
  default = false
}

variable "string" {
  type = string
}

variable "string_with_default" {
  type    = string
  default = "default_value"
}

variable "sensitive_string" {
  type      = string
  sensitive = true
}

variable "number" {
  type = number
}

variable "number_with_default" {
  type    = number
  default = 99
}

variable "list_of_bool" {
  type = list(bool)
}

variable "list_of_bool_with_default" {
  type    = list(bool)
  default = [false, true, false]
}

variable "list_of_string" {
  type = list(string)
}

variable "list_of_string_with_default" {
  type    = list(string)
  default = ["foo", "bar", "baz"]
}

variable "list_of_number" {
  type = list(number)
}

variable "list_of_number_with_default" {
  type    = list(number)
  default = [98, 99, 100]
}

variable "things" {
  type = list(
    object(
      {
        foo = list(number)
      }
    )
  )
}

variable "container" {
  type = object(
    {
      foo = string
      bar = object(
        {
          baz = string
          qux = list(
            object(
              {
                bing = string
                bong = number
              }
            )
          )
        }
      )
    }
  )
}


variable "optional_list" {
  type = optional(list(number), [])
}
