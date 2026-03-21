import { useFormContext } from "react-hook-form"

interface InputFieldProps extends React.InputHTMLAttributes<HTMLInputElement> {
  name: string
  label: string
}

export const InputField = ({ name, label, ...props }: InputFieldProps) => {
  const {
    register,
    formState: { errors },
  } = useFormContext()
  const error = errors[name]?.message as string | undefined

  return (
    <div className="mb-4 flex flex-col space-y-1">
      <label htmlFor={name} className="text-sm font-medium text-gray-700">
        {label}
      </label>
      <input
        id={name}
        {...register(name)}
        {...props}
        className={`rounded-md border px-3 py-2 outline-none focus:ring-2 ${
          error
            ? "border-red-500 focus:ring-red-200"
            : "border-gray-300 focus:ring-blue-200"
        }`}
      />
      {error && <span className="text-xs text-red-500">{error}</span>}
    </div>
  )
}
