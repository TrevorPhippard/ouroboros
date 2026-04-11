import { zodResolver } from "@hookform/resolvers/zod"
import {
  useForm,
  FormProvider,
  UseFormProps,
  FieldValues,
} from "react-hook-form"
import { ZodSchema } from "zod"

interface FormProps<TFormValues extends FieldValues> extends Omit<
  UseFormProps<TFormValues>,
  "resolver"
> {
  schema: ZodSchema<TFormValues>
  onSubmit: (data: TFormValues) => void
  children: React.ReactNode
  className?: string
}

export const Form = <TFormValues extends FieldValues>({
  schema,
  onSubmit,
  children,
  className,
  ...formOptions
}: FormProps<TFormValues>) => {
  const methods = useForm<TFormValues>({
    ...formOptions,
    resolver: zodResolver(schema),
  })

  return (
    <FormProvider {...methods}>
      <form onSubmit={methods.handleSubmit(onSubmit)} className={className}>
        {children}
      </form>
    </FormProvider>
  )
}
