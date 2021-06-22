import React, { FC, useState, useRef } from 'react'
import { TextInput } from '../../src/components'
import { PencilAltIcon, SaveIcon, XCircleIcon } from '@heroicons/react/outline'
import { Exact, AppInput } from '../../src/generated/graphql'
import { UseMutationResponse } from 'urql'
import c from 'classnames'

type Props = {
	useSet: () => UseMutationResponse<any, Exact<{
		input: AppInput
	}>>
	finished?: () => void
	value: string
	appName: string
	keyName: keyof AppInput
	className?: string
}

export const TextInputForm: FC<Props> = ({ appName, keyName, useSet, finished, value, className }) => {
	const [editing, setEditing] = useState<boolean>(false)
	const [loading, setLoading] = useState<boolean>(false)
	const valRef = useRef(null)
	const [setResult, set] = useSet()
	const err = setResult.error

	const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
		event.preventDefault()
		setLoading(true)

		const result = await set({
			input: {
				name:      appName,
				[keyName]: valRef.current.value,
			},
		})

		if (!result.error) {
			setEditing(false)
			setLoading(false)
			finished?.()
		}
	}

	return 	(
		<>
			{err &&
				<div>{err.message}</div>}

			<form onSubmit={onSubmit} className={c('flex', { 'opacity-10': loading })}>
				{editing
					? <TextInput
						name="value"
						disabled={(!editing) || loading}
						required
						ref={valRef}
						defaultValue={value}
						placeholder="value"
						className={className}
					/>
					: <span>{value}</span>
				}

				{editing
					? <XCircleIcon className="w-5 select-none" onClick={() => { setEditing(false); finished?.(); valRef.current.value = value }}/>
					: <PencilAltIcon className="w-5 select-none" onClick={() => setEditing(true)} />}

				{editing
					&& <button type="submit"><SaveIcon className="w-5" /></button>}
			</form>
		</>
	)
}
