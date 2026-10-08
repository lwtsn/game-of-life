import { create } from '@bufbuild/protobuf'
import { RulesSchema } from '../gen/life/v1/rules_pb.js'

const storageKey = 'gol.session'
const sessionPattern = new RegExp(create(RulesSchema).sessionPattern)

export function sessionId(): string {
  const existing = window.localStorage.getItem(storageKey)
  if (existing && sessionPattern.test(existing)) return existing

  const id = crypto.randomUUID()
  window.localStorage.setItem(storageKey, id)
  return id
}
