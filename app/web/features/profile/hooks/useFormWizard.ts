// src/hooks/useFormWizard.ts
import { useState } from "react"

export const useFormWizard = <TData>(
  stepsCount: number,
  initialData: Partial<TData> = {}
) => {
  const [currentStep, setCurrentStep] = useState(0)
  const [formData, setFormData] = useState<Partial<TData>>(initialData)

  const nextStep = (stepData: Partial<TData>) => {
    setFormData((prev) => ({ ...prev, ...stepData }))
    if (currentStep < stepsCount - 1) setCurrentStep((prev) => prev + 1)
  }

  const prevStep = () => {
    if (currentStep > 0) setCurrentStep((prev) => prev - 1)
  }

  return {
    currentStep,
    formData,
    nextStep,
    prevStep,
    isLastStep: currentStep === stepsCount - 1,
  }
}
