import {useSessionTerminalRetrieval} from '@/components/session/session-retrieval-hook';
import {IApp, ISession} from '@polo/common/state/models';
import {observer} from 'mobx-react-lite';
import React from 'react';

type TProps = {
	app: IApp;
	session: ISession;
	onSessionFail: () => void;
};

export const SessionTerminalContainer = observer((props: TProps) => {
	const containerRef = React.useRef<HTMLDivElement>();
	useSessionTerminalRetrieval(props.session, containerRef, props.app.failures.retrieveFailedSession, props.onSessionFail);

	return <div
		ref={containerRef}
		className={`w-full lg:m-2 lg:mt-5 flex-grow mt-10 mb-10 lg:mb-36 min-w-0 flex-1`}
		id={`terminal`}>

	</div>;
});