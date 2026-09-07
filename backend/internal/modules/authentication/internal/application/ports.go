package application

import "context"

// ITokenVerifierは、生のaccess tokenを検証して外部認証identityを返すinterfaceである。
// 具体的なJWT検証方法には依存せず、Supabase adapterなどから実装を注入する。
type ITokenVerifier interface {
	Verify(context.Context, string) (Identity, error)
}

// TokenVerifierFuncは、関数をITokenVerifierとして利用できるようにするadapterである。
type TokenVerifierFunc func(context.Context, string) (Identity, error)

func (f TokenVerifierFunc) Verify(ctx context.Context, token string) (Identity, error) {
	return f(ctx, token)
}
