log := (str => (
    out(str)
    out('
')
))

log2 := (str => (
    out(str)
    out('
')
    'log2 text'
))('hilog')

log('what wow')

log(log2)

kl := [5, 4, 3, 2, 1].2
ol := {
    ('te' + '-st'): 'magic'
}.('te-st')

log(ol)
log(string(kl))
