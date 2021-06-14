/* eslint-disable react/prop-types */
import { useRouter } from 'next/dist/client/router'
import React, { useRef } from 'react'
import { Card, Heading, TextInput, TextInputGroup, ExcludedResources } from '../../src/components'
import { throwError } from '../../src/throwError'
import { TrashIcon } from '@heroicons/react/outline'
import {
	useGetFullAppQuery,
	useDestroyAppMutation,
	useCreateReviewAppMutation,
	useSetExcludedResourceMutation,
	useSetAppVarMutation, useUnsetAppVarMutation,
	useSetAppEnvMutation, useUnsetAppEnvMutation,
} from '../../src/generated/graphql'


const CreateForm = ({ app }) => {
	const [{ error, fetching }, create] = useCreateReviewAppMutation()
	const branchNameRef = useRef(null)

	const submit = (event: React.FormEvent<HTMLFormElement>) => {
		event.preventDefault()

		create({
			input: {
				name:       app.name,
				branchName: branchNameRef.current.value,
			},
		})
	}

	return <form onSubmit={submit}>
		{error && <div className="bg-red-700 text-white border-red-700 p-2">
			{error.message}
		</div>}
		<TextInput name="branchName" ref={branchNameRef} placeholder="branch name" required disabled={fetching} />
		<button type="submit" className="rounded-sm p-1 underline" disabled={fetching}>Create</button>
	</form>
}

const ShowApp = () => {
	const router = useRouter()
	const id = router.query.id as string
	const [{ data: { getApp: app } }] = throwError(useGetFullAppQuery({ variables: { name: id } }))
	const [{ error: destroyAppError }, destroyApp] = useDestroyAppMutation()
	const hostname = `https://${app.name}.staging.freshlyservices.net/`

	return <div>
		<section className="p-3 mb-2">
			<h1 className="text-2xl">{app.name}</h1>
			<small>
				<a href={hostname} target="_blank" rel="noreferrer">{hostname}</a>
			</small>
		</section>

		<section>
			<Card className="mb-2">
				<h2 className="border-b-2">YAML Interpolation Vars</h2>
				<TextInputGroup
					vars={app.vars} appName={app.name}
					useSet={useSetAppVarMutation}
					useUnset={useUnsetAppVarMutation}
				/>
			</Card>

			<Card className="mb-2">
				<h2 className="border-b-2"> Environment Variables </h2>
				<TextInputGroup
					vars={app.env} appName={app.name}
					useSet={useSetAppEnvMutation}
					useUnset={useUnsetAppEnvMutation}
				/>
			</Card>
		</section>

		{app.reviewApp || <>
			<Card className="mb-2">
				<Heading>Create a review app</Heading>
				<CreateForm app={app} />
				<Heading>Review apps</Heading>
				{destroyAppError && <div className="bg-red-700 text-white border-red-700 p-2">
					{destroyAppError.message}
				</div>}

				{app.reviewApps && app.reviewApps.map(reviewApp =>
					<div key={reviewApp.name}>
						<a href={`/tuber/apps/${reviewApp.name}`}>{reviewApp.name}</a>
						<TrashIcon className="w-5" onClick={() => destroyApp({ input: { name: reviewApp.name } })}/>
					</div>,
				)}
			</Card>
		</>}

		<ExcludedResources
			appName={app.name}
			resources={app.reviewAppsConfig.excludedResources}
			useSet={useSetExcludedResourceMutation}
		/>
	</div>
}

export default ShowApp
